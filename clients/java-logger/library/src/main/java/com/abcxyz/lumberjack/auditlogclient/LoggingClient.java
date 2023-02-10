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
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogBackend;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogMutator;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogValidator;
import com.abcxyz.lumberjack.auditlogclient.processor.PreconditionFailedException;
import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;
import java.util.List;
import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;

/** Logging client for lumberjack audit logging */
@Getter(AccessLevel.PROTECTED)
@AllArgsConstructor
@Slf4j
public class LoggingClient {
  private final List<LogValidator> validators;
  private final List<LogMutator> mutators;
  private final List<LogBackend> backends;
  private final AuditLoggingConfiguration config;

  /**
   * Iterates through all the log processors for a client sequentially and calls their {@link
   * LogProcessor#process(AuditLogRequest)}
   *
   * @param auditLogRequest log request
   */
  public void log(AuditLogRequest auditLogRequest) throws LogProcessingException {
    // Override config's log mode if the request has explicitly specified log mode. We also want to
    // ensure that the log mode used here is passed on to the server, so if the log mode is missing
    // from the request, we add the config log mode to the request.
    LogMode logMode;
    if (auditLogRequest.getMode() == null
        || auditLogRequest.getMode().equals(LogMode.LOG_MODE_UNSPECIFIED)) {
      logMode = config.getLogMode();
      auditLogRequest = auditLogRequest.toBuilder().setMode(logMode).build();
    } else {
      logMode = auditLogRequest.getMode();
    }

    try {
      for (LogProcessor processor : validators) {
        auditLogRequest = processor.process(auditLogRequest);
      }
      for (LogProcessor processor : mutators) {
        auditLogRequest = processor.process(auditLogRequest);
      }
      for (LogProcessor processor : backends) {
        auditLogRequest = processor.process(auditLogRequest);
      }
    } catch (PreconditionFailedException e) {
      log.warn("Stopped log request processing.", e);
    } catch (Exception e) { // TODO(#157): Should we swallow throwable?

      if (ConfigUtils.shouldFailClose(logMode)) {
        throw new LogProcessingException("Failed to audit log.", e);
      } else {
        log.error("Failed to audit log; continuing without audit logging.", e);
      }
    }
  }
}
