/*
 * Copyright 2023 Lumberjack authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.abcxyz.lumberjack.auditlogclient;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.config.Selector;
import com.abcxyz.lumberjack.auditlogclient.exceptions.AuthorizationException;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.google.inject.Inject;
import com.google.logging.v2.LogEntryOperation;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.MessageOrBuilder;
import com.google.protobuf.Struct;
import com.google.protobuf.Timestamp;
import com.google.protobuf.Value;
import com.google.protobuf.util.JsonFormat;
import com.google.rpc.Code;
import com.google.rpc.Status;
import io.grpc.Context;
import io.grpc.Contexts;
import io.grpc.ForwardingServerCall.SimpleForwardingServerCall;
import io.grpc.ForwardingServerCallListener;
import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCall.Listener;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;
import io.grpc.StatusRuntimeException;
import java.time.Clock;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Deque;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.concurrent.ConcurrentLinkedDeque;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/** This is intended to allow automatic audit logging for calls from a wrapped server. */
@RequiredArgsConstructor(onConstructor = @__({@Inject}))
@Slf4j
public class AuditLoggingServerInterceptor<ReqT extends Message> implements ServerInterceptor {
  public static final String JUSTIFICATION_TOKEN_HEADER_KEY = "justification-token";
  public static final Context.Key<AuditLog.Builder> AUDIT_LOG_CTX_KEY = Context.key("audit-log");
  public static final String UNSPECIFIED_RESORCE = "GRPC_STREAM_RESOURCE_NAME_PLACEHOLDER";

  private static final Metadata.Key<String> JUSTIFICATION_METADATA_KEY =
      Metadata.Key.of(JUSTIFICATION_TOKEN_HEADER_KEY, Metadata.ASCII_STRING_MARSHALLER);

  /**
   * Keeps track of the relevant selectors for specific methods. As the selectors that are relevant
   * for each method doesn't change, it is more efficient to keep track of them after determining
   * once whether they are applicable, rather than iterating over each selector on every single
   * method call. This is used in the getRelevantSelector method.
   */
  private final Map<String, Optional<Selector>> memo = new HashMap<>();

  private final AuditLoggingConfiguration auditLoggingConfiguration;
  private final LoggingClient client;
  private final Clock clock;

  @Override
  public <ReqT, RespT> Listener<ReqT> interceptCall(
      ServerCall<ReqT, RespT> call, Metadata headers, ServerCallHandler<ReqT, RespT> next) {
    String methodName = call.getMethodDescriptor().getFullMethodName();
    Optional<Selector> selectorOption = getRelevantSelector(methodName);
    if (selectorOption.isEmpty()) {
      log.info("No selector found for method {}", methodName);
      return next.startCall(call, headers);
    }
    Selector selector = selectorOption.get();

    Optional<String> principal = Optional.empty();
    try {
      principal = auditLoggingConfiguration.getSecurityContext().getPrincipal(headers);
    } catch (AuthorizationException e) {
      log.warn("Exception while trying to determine principal.");
      if (ConfigUtils.shouldFailClose(auditLoggingConfiguration.getLogMode())) {
        throw new IllegalStateException("Unable to determine principal.", e);
      } else {
        log.error("Principal was unable to be determined, continuing without audit logging.", e);
        next.startCall(call, headers);
      }
    }

    AuditLog.Builder logBuilder = AuditLog.newBuilder();
    String fullMethodName = call.getMethodDescriptor().getFullMethodName();
    logBuilder.setMethodName(fullMethodName);
    logBuilder.setResourceName(UNSPECIFIED_RESORCE);
    // if the client has multiple streaming uploads before there is a response,
    logBuilder.setServiceName(fullMethodName.split("/")[0]);

    if (principal.isPresent()) {
      logBuilder.setAuthenticationInfo(
          AuthenticationInfo.newBuilder().setPrincipalEmail(principal.get()).build());
    } else {
      log.info("Unable to determine principal for request.");
      next.startCall(call, headers);
    }

    final Struct auditLogRequestContext = getAuditLogRequestContext(headers);

    LogEntryOperation logEntryOperation =
        LogEntryOperation.newBuilder()
            .setId(UUID.randomUUID().toString())
            .setProducer(fullMethodName)
            .build();

    // Add the builder into the context, this makes it available to the server code.
    Context ctx = Context.current().withValue(AUDIT_LOG_CTX_KEY, logBuilder);

    // Create a final logBuilder to make lambda happy.
    final AuditLog.Builder logBuilderFinal = logBuilder;
    // Deques allow for addition/removal at both ends. We use this to keep responses
    // until it is time to log them.
    Deque<ReqT> unloggedRequests = new ConcurrentLinkedDeque<>();
    ServerCall<ReqT, RespT> delegateCall =
        new SimpleForwardingServerCall<ReqT, RespT>(call) {
          @Override
          public void sendMessage(RespT message) {
            // Mewest message. Returns null if empty.
            ReqT unloggedRequest = unloggedRequests.pollLast();
            auditLog(
                selector,
                auditLogRequestContext,
                unloggedRequest,
                message,
                logBuilderFinal,
                logEntryOperation);
            super.sendMessage(message);
          }

          // Always handle non-OK status in the close call.
          // Exceptions don't always propagate on the same path. It's most effective to capture the
          // error code in the close call.
          @Override
          public void close(io.grpc.Status status, Metadata trailers) {
            if (status != io.grpc.Status.OK) {
              // Audit logs expect an rpc code, however this exception is grpc specific. We have to
              // convert from one to the other.
              Code code = Code.forNumber(status.getCode().value());
              ReqT unloggedRequest = unloggedRequests.pollFirst(); // try to get the last request
              logBuilder.setStatus(
                  Status.newBuilder().setCode(code.getNumber()).setMessage(code.name()).build());
              auditLog(
                  selector,
                  auditLogRequestContext,
                  unloggedRequest,
                  null,
                  logBuilder,
                  logEntryOperation);
            }
            super.close(status, trailers);
          }
        };
    ServerCall.Listener<ReqT> delegate = Contexts.interceptCall(ctx, delegateCall, headers, next);

    // we keep a running queue of unlogged requests. This is intended to only hold a single one, but
    // it is possible for more than one request to end up in the queue. This allows us to associate
    // a request with each without double logging responses. If the case occurs where multiple
    // requests come in before a response occurs, we log all but the last request, then log the last
    // request with the response. The timing is all best-effort, and no guarantees are made on
    // ordering. It is also possible in the server streaming case (where multiple responses are
    // returned for a single request) that only the first response will have an associated request.
    return new ForwardingServerCallListener.SimpleForwardingServerCallListener<ReqT>(delegate) {
      @Override
      public void onMessage(ReqT message) {
        while (!unloggedRequests.isEmpty()) {
          ReqT unloggedRequest = unloggedRequests.pollFirst(); // oldest message
          // between the isEmpty() and the poll, another thread could have grabbed it,
          // so we need to check for null.
          if (unloggedRequest != null) {
            auditLog(
                selector,
                auditLogRequestContext,
                unloggedRequest,
                null,
                logBuilderFinal,
                logEntryOperation);
          }
        }
        unloggedRequests.add(message);

        try {
          super.onMessage(message);
        } catch (Exception e) {
          closeWithException(e);
        }
      }

      @Override
      public void onHalfClose() {
        try {
          super.onHalfClose();
        } catch (Exception e) {
          closeWithException(e);
        }
      }

      @Override
      public void onComplete() {
        try {
          super.onComplete();
        } catch (Exception e) {
          closeWithException(e);
        }
      }

      @Override
      public void onCancel() {
        try {
          super.onCancel();
        } catch (Exception e) {
          closeWithException(e);
        }
      }

      // Explicitly call.close() for any exceptions we run into on these events.
      // This allows us to leverage call.close() to be the single place to write
      // the last audit log with the error response code.
      private void closeWithException(Exception e) {
        StatusRuntimeException t;
        Metadata metadata = new Metadata();

        if (e instanceof StatusRuntimeException) {
          t = (StatusRuntimeException) e;
          if (t.getTrailers() != null) {
            metadata = t.getTrailers();
          }
        } else {
          // Treat other exceptions as UNKNOWN status.
          t = new StatusRuntimeException(io.grpc.Status.UNKNOWN);
        }

        delegateCall.close(t.getStatus(), metadata);
      }
    };
  }

  <ReqT, RespT> void auditLog(
      Selector selector,
      Struct auditLogRequestContext,
      ReqT request,
      RespT response,
      AuditLog.Builder logBuilder,
      LogEntryOperation logEntryOperation) {
    AuditLog.Builder logBuilderCopy = logBuilder.build().toBuilder();
    if (selector.getDirective().shouldLogResponse() && response != null) {
      logBuilderCopy.setResponse(messageToStruct(response));
    }
    if (selector.getDirective().shouldLogRequest() && request != null) {
      logBuilderCopy.setRequest(messageToStruct(request));
    }

    AuditLogRequest.Builder builder = AuditLogRequest.newBuilder();
    builder.setPayload(logBuilderCopy.build());
    builder.setType(selector.getLogType());
    builder.setOperation(logEntryOperation);
    Instant now = clock.instant();
    builder.setTimestamp(
        Timestamp.newBuilder().setSeconds(now.getEpochSecond()).setNanos(now.getNano()));
    builder.setContext(auditLogRequestContext);

    try {
      log.info("Audit logging...");
      client.log(builder.build());
    } catch (LogProcessingException e) {
      throw new RuntimeException(e);
    }
  }

  /**
   * Converts a proto message of unknown type to a proto struct. In order to do this, the method
   * first converts the message to json, and then from json to a protobuf struct.
   *
   * <p>TODO: This may not be the most efficient way , and it would be beneficial to find a solution
   * that avoids the middleman, and can convert directly from MessageOrBuilder to a Struct.
   */
  <ReqT> Struct messageToStruct(ReqT message) {
    if (message instanceof MessageOrBuilder) {
      Struct.Builder structBuilder = Struct.newBuilder();
      try {
        String jsonString = JsonFormat.printer().print((MessageOrBuilder) message);
        JsonFormat.parser().merge(jsonString, structBuilder);
        return structBuilder.build();
      } catch (InvalidProtocolBufferException e) {
        throw new RuntimeException(e);
      }
    } else {
      throw new IllegalArgumentException("Not a Protobuf Message: " + message.toString());
    }
  }

  /**
   * Converts a list of proto messages to a human-readable string, and then puts that string into a
   * struct for use when audit logging.
   *
   * <p>TODO: this may not be the most optimal if we want to consume and do processing later on this
   * information. consider changing to a format that would be more conducive to later consumption
   */
  <ReqT> Struct messagesToStruct(List<ReqT> messages) {
    List<String> messageStrings = new ArrayList<>();
    for (ReqT message : messages) {
      messageStrings.add(message.toString());
    }
    Struct.Builder structBuilder = Struct.newBuilder();
    String formattedList = messageStrings.toString().replace("\n", "");

    structBuilder.putFields(
        "request_list", Value.newBuilder().setStringValue(formattedList).build());
    return structBuilder.build();
  }

  Optional<Selector> getRelevantSelector(String methodIdentifier) {
    if (memo.containsKey(methodIdentifier)) {
      return memo.get(methodIdentifier);
    }
    Optional<Selector> mostApplicableSelector =
        Selector.returnMostRelevant(methodIdentifier, auditLoggingConfiguration.getRules());

    // thread-safe way to update memo
    memo.putIfAbsent(methodIdentifier, mostApplicableSelector);
    return mostApplicableSelector;
  }

  /**
   * Get the justification token out of the headers, and create a struct of the correct format for
   * use in audit logging context.
   */
  Struct getAuditLogRequestContext(Metadata headers) {
    String jvsToken = headers.get(JUSTIFICATION_METADATA_KEY);
    if (jvsToken != null && !jvsToken.isEmpty()) {
      return Struct.newBuilder()
          .putFields(
              JUSTIFICATION_TOKEN_HEADER_KEY, Value.newBuilder().setStringValue(jvsToken).build())
          .build();
    } else {
      return Struct.getDefaultInstance();
    }
  }
}
