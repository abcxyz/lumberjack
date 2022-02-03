/*
 * Copyright 2021 Lumberjack authors (see AUTHORS file)
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
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.google.inject.Inject;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.MessageOrBuilder;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import com.google.protobuf.util.JsonFormat;
import io.grpc.Context;
import io.grpc.Contexts;
import io.grpc.ForwardingServerCall.SimpleForwardingServerCall;
import io.grpc.ForwardingServerCallListener;
import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCall.Listener;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.logging.Logger;
import lombok.RequiredArgsConstructor;

/** This is intended to allow automatic audit logging for calls from a wrapped server. */
@RequiredArgsConstructor(onConstructor = @__({@Inject}))
public class AuditLoggingServerInterceptor<ReqT extends Message> implements ServerInterceptor {
  private static final Logger log = Logger.getLogger(AuditLoggingServerInterceptor.class.getName());
  public static final Context.Key<AuditLog.Builder> AUDIT_LOG_CTX_KEY = Context.key("audit-log");

  /**
   * Keeps track of the relevant selectors for specific methods. As the selectors that are relevant
   * for each method doesn't change, it is more efficient to keep track of them after determining
   * once whether they are applicable, rather than iterating over each selector on every single
   * method call. This is used in the getRelevantSelector method.
   */
  private final Map<String, Optional<Selector>> memo = new HashMap<>();

  private final AuditLoggingConfiguration auditLoggingConfiguration;
  private final LoggingClient client;

  @Override
  public <ReqT, RespT> Listener<ReqT> interceptCall(
      ServerCall<ReqT, RespT> call, Metadata headers, ServerCallHandler<ReqT, RespT> next) {
    String methodName = call.getMethodDescriptor().getFullMethodName();
    Optional<Selector> selectorOption = getRelevantSelector(methodName);
    if (selectorOption.isEmpty()) {
      log.info("No selector found for method {}" + methodName);
      return next.startCall(call, headers);
    }
    Selector selector = selectorOption.get();

    Optional<String> principal = Optional.empty();
    try {
      principal = auditLoggingConfiguration.getSecurityContext().getPrincipal(headers);
    } catch (AuthorizationException e) {
      log.info("Exception while trying to determine principal..");
      next.startCall(call, headers);
    }

    AuditLog.Builder logBuilder = AuditLog.newBuilder();
    String fullMethodName = call.getMethodDescriptor().getFullMethodName();
    logBuilder.setMethodName(fullMethodName);
    logBuilder.setServiceName(fullMethodName.split("/")[0]);

    if (principal.isPresent()) {
      logBuilder.setAuthenticationInfo(
          AuthenticationInfo.newBuilder().setPrincipalEmail(principal.get()).build());
    } else {
      log.info("Unable to determine principal for request.");
      next.startCall(call, headers);
    }

    Context ctx = Context.current().withValue(AUDIT_LOG_CTX_KEY, logBuilder);

    // Add the builder into the context, this makes it available to the server code.
    ServerCall.Listener<ReqT> delegate =
        Contexts.interceptCall(
            ctx,
            new SimpleForwardingServerCall<ReqT, RespT>(call) {
              @Override
              public void sendMessage(RespT message) {
                if (selector.getDirective().shouldLogResponse()) {
                  Struct struct = messageToStruct(message);
                  logBuilder.setResponse(struct);
                }

                AuditLogRequest.Builder builder = AuditLogRequest.newBuilder();
                builder.setPayload(logBuilder.build());
                builder.setType(selector.getLogType());

                try {
                  log.info("Audit log: " + builder.build().toString());
                  client.log(builder.build());
                } catch (LogProcessingException e) {
                  throw new RuntimeException(e);
                }

                super.sendMessage(message);
              }
            },
            headers,
            next);

    // If the client is streaming, we will get multiple requests.
    List<ReqT> requests = new ArrayList<>();
    return new ForwardingServerCallListener.SimpleForwardingServerCallListener<ReqT>(delegate) {
      @Override
      public void onMessage(ReqT message) {
        if (selector.getDirective().shouldLogRequest()) {
          requests.add(message);
          logBuilder.setRequest(
              requests.size() > 1 ? messagesToStruct(requests) : messageToStruct(message));
        }
        super.onMessage(message);
      }
    };
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
}
