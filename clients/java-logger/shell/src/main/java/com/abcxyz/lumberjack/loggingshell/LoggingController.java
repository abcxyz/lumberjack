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
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import lombok.RequiredArgsConstructor;
import lombok.extern.log4j.Log4j2;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestAttribute;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

/** Endpoints for the shell app that imports/uses the Audit Logging client library. */
@Log4j2
@RequiredArgsConstructor
@RestController
public class LoggingController {
  static final String TRACE_ID_PARAMETER_KEY = "trace_id";
  private static final String SERVICE_NAME = "java-shell-app";

  private final LoggingClient loggingClient;

  @GetMapping
  @ResponseStatus(value = HttpStatus.OK)
  void loggingShell(
      @RequestParam(value = TRACE_ID_PARAMETER_KEY) String traceId,
      @RequestAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY) String userEmail)
      throws LogProcessingException {
    AuditLogRequest record =
        AuditLogRequest.newBuilder()
            .setPayload(
                AuditLog.newBuilder()
                    .setServiceName(SERVICE_NAME)
                    .setResourceName(traceId)
                    .setMethodName("loggingShell")
                    .setAuthenticationInfo(
                        AuthenticationInfo.newBuilder().setPrincipalEmail(userEmail).build()))
            .setType(LogType.DATA_ACCESS)
            .putLabels(TRACE_ID_PARAMETER_KEY, traceId)
            .build();
    loggingClient.log(record);
    log.info("Logged successfully with trace id: {} for user: {}", traceId, userEmail);
  }
}
