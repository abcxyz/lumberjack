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
import com.google.protobuf.Timestamp;
import java.time.Clock;
import java.time.Instant;
import javax.annotation.PreDestroy;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.web.bind.annotation.*;

/** Endpoints for the shell app that imports/uses the Audit Logging client library. */
@RestController
@RequestMapping(path = "/")
public class LoggingController {
  static final String JUSTIFICATION_TOKEN_HEADER_NAME = "justification-token";
  static final String TRACE_ID_PARAMETER_KEY = "trace_id";

  private static final Logger log = LoggerFactory.getLogger(LoggingController.class);
  private static final String SERVICE_NAME = "java-shell-app";

  private final LoggingClient loggingClient;

  private final Clock clock;

  public LoggingController(LoggingClient loggingClient, Clock clock) {
    this.loggingClient = loggingClient;
    this.clock = clock;
  }

  @PreDestroy
  public void destroy() {
    loggingClient.close();
  }

  @GetMapping
  public void loggingShell(
      @RequestParam(value = TRACE_ID_PARAMETER_KEY) String traceId,
      @RequestAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY) String userEmail,
      @RequestHeader(JUSTIFICATION_TOKEN_HEADER_NAME) String jvsToken)
      throws LogProcessingException {
    Instant now = clock.instant();
    AuditLogRequest record =
        AuditLogRequest.newBuilder()
            .setTimestamp(
                Timestamp.newBuilder().setSeconds(now.getEpochSecond()).setNanos(now.getNano()))
            .setPayload(
                AuditLog.newBuilder()
                    .setServiceName(SERVICE_NAME)
                    .setResourceName(traceId)
                    .setMethodName("loggingShell")
                    .setAuthenticationInfo(
                        AuthenticationInfo.newBuilder().setPrincipalEmail(userEmail).build()))
            .setType(LogType.DATA_ACCESS)
            .putLabels(TRACE_ID_PARAMETER_KEY, traceId)
            .setJustificationToken(jvsToken)
            .build();
    loggingClient.log(record);
    log.info(
        "Logged successfully with trace id: {} justification tokenL {} for user: {}",
        traceId,
        jvsToken,
        userEmail);
  }
}
