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
import com.abcxyz.lumberjack.auditlogclient.config.JwtSpecification;
import com.abcxyz.lumberjack.auditlogclient.config.Selector;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.auth0.jwt.JWT;
import com.auth0.jwt.exceptions.JWTDecodeException;
import com.auth0.jwt.interfaces.Claim;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.google.inject.Inject;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Message;
import com.google.protobuf.MessageOrBuilder;
import com.google.protobuf.Struct;
import com.google.protobuf.util.JsonFormat;
import io.grpc.ForwardingServerCall.SimpleForwardingServerCall;
import io.grpc.ForwardingServerCallListener;
import io.grpc.Metadata;
import io.grpc.ServerCall;
import io.grpc.ServerCall.Listener;
import io.grpc.ServerCallHandler;
import io.grpc.ServerInterceptor;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/** This is intended to allow automatic audit logging for calls from a wrapped server. */
@Slf4j
@RequiredArgsConstructor(onConstructor = @__({@Inject}))
public class AuditLoggingServerInterceptor<ReqT extends Message> implements ServerInterceptor {
  // TODO: Move security-specific logic out and add easier extensibility.
  private static final String EMAIL_KEY = "email";

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
      log.debug("No selector found for method {}", methodName);
      return next.startCall(call, headers);
    }
    Selector selector = selectorOption.get();

    Optional<String> principal = getPrincipalFromJwt(headers);

    AuditLog.Builder logBuilder = AuditLog.newBuilder();
    String fullMethodName = call.getMethodDescriptor().getFullMethodName();
    logBuilder.setMethodName(fullMethodName);
    logBuilder.setServiceName(fullMethodName.split("/")[0]);

    if (principal.isPresent()) {
      logBuilder.setAuthenticationInfo(
          AuthenticationInfo.newBuilder().setPrincipalEmail(principal.get()).build());
    } else {
      log.debug("Unable to determine principal for request.");
    }

    Listener<ReqT> delegate =
        next.startCall(
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
            headers);

    return new ForwardingServerCallListener.SimpleForwardingServerCallListener<ReqT>(delegate) {
      @Override
      public void onMessage(ReqT message) {
        if (selector.getDirective().shouldLogRequest()) {
          Struct struct = messageToStruct(message);
          logBuilder.setRequest(struct);
        }

        super.onMessage(message);
      }
    };
  }

  /**
   * TODO: currently no handling is implemented for JWKS checks.
   */
  Optional<String> getPrincipalFromJwt(Metadata headers) {
    List<JwtSpecification> jwtSpecifications =
        auditLoggingConfiguration.getSecurityContext().getJwtSpecifications();
    for (JwtSpecification jwtSpecification : jwtSpecifications) {
      Metadata.Key<String> metadataKey =
          Metadata.Key.of(jwtSpecification.getKey(), Metadata.ASCII_STRING_MARSHALLER);
      if (!headers.containsKey(metadataKey)) {
        continue;
      }
      String idToken = headers.get(metadataKey);
      if (idToken.startsWith(jwtSpecification.getPrefix())) {
        idToken = idToken.substring(jwtSpecification.getPrefix().length());
      }
      try {
        DecodedJWT jwt = JWT.decode(idToken);
        Map<String, Claim> claims = jwt.getClaims();
        if (!claims.containsKey(EMAIL_KEY)) {
          continue;
        }
        String principal = claims.get(EMAIL_KEY).asString();
        log.info("Found JWT key {} with email {}", jwtSpecification.getKey(), principal);
        return Optional.of(principal);
      } catch (JWTDecodeException e) {
        // invalid token
        throw new RuntimeException(e);
      }
    }

    return Optional.empty();
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
