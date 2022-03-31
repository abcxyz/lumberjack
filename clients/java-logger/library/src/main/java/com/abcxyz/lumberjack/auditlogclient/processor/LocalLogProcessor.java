/*
 * Copyright 2022 Lumberjack authors (see AUTHORS file)
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
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;

/** Logs the {@link AuditLogRequest} using the standard logger. */
@Slf4j
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class LocalLogProcessor implements LogBackend {

  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) {
    try {
      String jsonString = JsonFormat.printer().omittingInsignificantWhitespace().print(auditLogRequest);
      log.info("Lumberjack log: {}", jsonString);
    } catch (InvalidProtocolBufferException e) {
      throw new RuntimeException(e);
    }
    return auditLogRequest;
  }
}
