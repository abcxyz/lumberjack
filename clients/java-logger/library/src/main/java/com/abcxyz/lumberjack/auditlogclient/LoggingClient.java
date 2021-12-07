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

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import java.util.List;
import lombok.AccessLevel;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor;

/**
 * Logging client for lumberjack audit logging
 */
@RequiredArgsConstructor
@Getter(AccessLevel.PROTECTED)
public class LoggingClient {

  private final List<LogProcessor> validators;
  private final List<LogProcessor> mutators;
  private final List<LogProcessor> backends;

  /**
   * Iterates through all the log processors for a client sequentially and calls
   * their {@link
   * LogProcessor#process(AuditLogRequest)}
   *
   * @param auditLogRequest log request
   */
  public void log(AuditLogRequest auditLogRequest) throws LogProcessingException {
    for (LogProcessor processor : validators) {
      auditLogRequest = processor.process(auditLogRequest);
    }
    for (LogProcessor processor : mutators) {
      auditLogRequest = processor.process(auditLogRequest);
    }
    for (LogProcessor processor : backends) {
      auditLogRequest = processor.process(auditLogRequest);
    }
  }
}
