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

package com.abcxyz.lumberjack.loggingshell;

import com.abcxyz.lumberjack.auditlogclient.LoggingClient;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogType;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import java.util.Map;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestAttribute;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

/** Endpoints for the shell app that imports/uses the Audit Logging client library. */
@Slf4j
@RequiredArgsConstructor
@RestController
public class LoggingController {
  static final String TRACE_ID_PARAMETER_KEY = "trace_id";
  private static final String SERVICE_NAME = "java-shell-app";
  private static final String METHOD_NAME = "abcxyz.LoggingShellApplication.GetLoggingShell";
  private static final String RESOURCE_NAME = "projects/LoggingShell";
  private static final String METADATA_FIELD_NAME = "data";
  private static final Map<String, String> METADATA_MAP =
      Map.of(
          "fieldA", "valueA",
          "fieldB", "valueB",
          "fieldC", "valueC");

  private final ObjectMapper objectMapper;
  private final LoggingClient loggingClient;

  @GetMapping
  @ResponseStatus(value = HttpStatus.OK)
  void loggingShell(
      @RequestParam(value = TRACE_ID_PARAMETER_KEY) String traceId,
      @RequestAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY) String userEmail)
      throws LogProcessingException, JsonProcessingException {
    AuditLogRequest record =
        AuditLogRequest.newBuilder()
            .setPayload(
                AuditLog.newBuilder()
                    .setServiceName(SERVICE_NAME)
                    .setMethodName(METHOD_NAME)
                    .setResourceName(RESOURCE_NAME)
                    .setMetadata(
                        Struct.newBuilder()
                            .putFields(
                                METADATA_FIELD_NAME,
                                Value.newBuilder()
                                    .setStringValue(objectMapper.writeValueAsString(METADATA_MAP))
                                    .build())
                            .build())
                    .setAuthenticationInfo(
                        AuthenticationInfo.newBuilder().setPrincipalEmail(userEmail).build()))
            .setType(LogType.DATA_ACCESS)
            .putLabels(TRACE_ID_PARAMETER_KEY, traceId)
            .build();
    loggingClient.log(record);
    log.info("Logged successfully with trace id: {} for user: {}", traceId, userEmail);
  }
}
