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

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogMutator;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.inject.Inject;
import lombok.AllArgsConstructor;

/**
 * Adds labels from the configuration to the audit log request. Does not overwrite existing labels.
 */
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class LabelProcessor implements LogMutator {
  private AuditLoggingConfiguration config;

  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) throws LogProcessingException {
    if (config.getLabels() == null || config.getLabels().isEmpty()) {
      // shortcut if there are no labels to add
      return auditLogRequest;
    }
    AuditLogRequest.Builder builder = auditLogRequest.toBuilder();
    config.getLabels().entrySet().stream()
        .filter(e -> !auditLogRequest.getLabelsMap().containsKey(e.getKey()))
        .forEach(e -> builder.putLabels(e.getKey(), e.getValue()));
    return builder.build();
  }
}
