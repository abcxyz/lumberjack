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

import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogBackend;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.inject.Inject;
import lombok.AllArgsConstructor;
import lombok.extern.java.Log;

/** Logs the {@link AuditLogRequest} using the standard logger.. */
@Log
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class LocalLogProcessor implements LogBackend {

  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) {
    // TODO: Do we want system.out or log here?
    log.info("Lumberjack audit log: " + auditLogRequest.toString());
    return auditLogRequest;
  }
}
