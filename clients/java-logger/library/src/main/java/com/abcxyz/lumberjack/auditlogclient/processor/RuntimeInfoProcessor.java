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

import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogMutator;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.cloud.audit.AuditLog;
import com.google.inject.Inject;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import javax.annotation.Nullable;
import lombok.AllArgsConstructor;

/**
 * RuntimeInfo is a processor that contains information about the application's runtime environment.
 */
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class RuntimeInfoProcessor implements LogMutator {
  @Nullable private final Value monitoredResource;

  /**
   * Process stores the application's GCP runtime information in the audit log request. More
   * specifically, in the Payload.Metadata under the key "originating_resource".
   *
   * @return modified {@link AuditLogRequest}
   */
  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) {
    if (monitoredResource == null) {
      return auditLogRequest;
    }
    // Add monitored resource to Payload.Metadata as JSON.
    if (!auditLogRequest.getPayload().hasMetadata()) {
      AuditLogRequest.Builder auditLogRequestToUpdate = auditLogRequest.toBuilder();
      AuditLog.Builder auditLogToUpdate = auditLogRequest.getPayload().toBuilder();
      auditLogToUpdate.setMetadata(Struct.newBuilder().build());
      auditLogRequest =
          auditLogRequestToUpdate.clearPayload().setPayload(auditLogToUpdate.build()).build();
    }

    // add new field with monitoredResource to existing metadata
    AuditLogRequest.Builder auditLogRequestToUpdate = auditLogRequest.toBuilder();
    AuditLog.Builder auditLogToUpdate = auditLogRequest.getPayload().toBuilder();
    Struct.Builder metadataToUpdate = auditLogRequest.getPayload().getMetadata().toBuilder();
    metadataToUpdate.putFields("originating_resource", monitoredResource);
    auditLogToUpdate.clearMetadata().setMetadata(metadataToUpdate.build());
    return auditLogRequestToUpdate.clearPayload().setPayload(auditLogToUpdate.build()).build();
  }
}
