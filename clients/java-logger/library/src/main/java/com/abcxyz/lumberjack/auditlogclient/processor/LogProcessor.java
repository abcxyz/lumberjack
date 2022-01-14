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

package com.abcxyz.lumberjack.auditlogclient.processor;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;

/**
 * This interface is defined to add different processors to process an {@link AuditLogRequest }
 *
 * <p>Example
 *
 * <ul>
 *   <li>validation processor to validate the {@link AuditLogRequest}auditLogRequest
 *   <li>processor to send the audit log to the GCP
 * </ul>
 */
public interface LogProcessor {
  /**
   * Performs the "processing logic" on the given {@link AuditLogRequest} and returns an new one
   * with updates if any.
   */
  AuditLogRequest process(AuditLogRequest auditLogRequest) throws LogProcessingException;

  interface LogValidator extends LogProcessor {}

  interface LogMutator extends LogProcessor {}

  interface LogBackend extends LogProcessor {}
}
