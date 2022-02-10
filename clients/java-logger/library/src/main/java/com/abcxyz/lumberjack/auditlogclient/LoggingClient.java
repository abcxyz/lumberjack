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
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import java.util.List;
import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.extern.java.Log;

/** Logging client for lumberjack audit logging */
@Getter(AccessLevel.PROTECTED)
@AllArgsConstructor
@Log
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
    } catch (Exception e) { // TODO: Should we swallow throwable?
      if (config.shouldFailClose()) {
        log.info("Log mode is fail close, raising up error.");
        throw e;
      } else {
        log.warning("Log mode isn't fail close, swallowing error: " + e.getMessage());
      }
    }
  }
}
