/*
 * Copyright 2022 Lumberjack authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package com.abcxyz.lumberjack.auditlogclient.processor;

import com.abcxyz.jvs.JvsClient;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.api.client.util.StringUtils;
import com.google.cloud.audit.AuditLog;
import com.auth0.jwt.interfaces.DecodedJWT;
import org.apache.commons.codec.binary.Base64;
import com.google.inject.Inject;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import com.google.protobuf.util.JsonFormat;
import lombok.AllArgsConstructor;

@AllArgsConstructor(onConstructor = @__({@Inject}))
public class JustificationProcessor {
  private static final String JUSTIFICATION_LOG_METADATA_KEY = "justification";
  private final JvsClient jvs;

  /**
   * Validates the given {@code jvsToken} and populates the {@link AuditLogRequest} with the
   * justification info.
   *
   * @param jvsToken A JWT produced by JVS.
   * @param auditLogRequest Audit log request to be processed.
   * @return The {@link AuditLogRequest} with justification filled.
   * @throws LogProcessingException when it fails to populate the justification info.
   */
  public AuditLogRequest process(String jvsToken, AuditLogRequest auditLogRequest)
      throws LogProcessingException {
    if (!auditLogRequest.hasPayload()) {
      throw new IllegalArgumentException("audit log request doesn't have payload");
    }

    AuditLogRequest.Builder auditLogRequestBuilder = auditLogRequest.toBuilder();
    AuditLog.Builder auditLogBuilder = auditLogRequest.getPayload().toBuilder();
    this.setLogJustification(jvsToken, auditLogBuilder);

    return auditLogRequestBuilder.setPayload(auditLogBuilder.build()).build();
  }

  /**
   * Validates the given {@code jvsToken} and sets the justification info in the given
   * {@code auditLogBuilder}.
   *
   * @param jvsToken A JWT produced by JVS.
   * @param auditLogBuilder Audit log builder.
   * @throws LogProcessingException when it fails to set the justification info.
   */
  public void setLogJustification(String jvsToken, AuditLog.Builder auditLogBuilder)
      throws LogProcessingException {
    try {
      DecodedJWT jwt = jvs.validateJWT(jvsToken);
      String jsonString = StringUtils.newStringUtf8(Base64.decodeBase64(jwt.getPayload()));
      Struct.Builder justificationStructBuilder = Struct.newBuilder();
      JsonFormat.parser().merge(jsonString, justificationStructBuilder);
      Struct justificationStruct = justificationStructBuilder.build();

      // Add monitored resource to Payload.Metadata as JSON.
      if (!auditLogBuilder.hasMetadata()) {
        auditLogBuilder.setMetadata(Struct.newBuilder().build());
      }

      // add new field with monitoredResource to existing metadata
      Struct.Builder metadataBuilder = auditLogBuilder.getMetadata().toBuilder();
      metadataBuilder.putFields(JUSTIFICATION_LOG_METADATA_KEY,
          Value.newBuilder().setStructValue(justificationStruct).build());
      auditLogBuilder.setMetadata(metadataBuilder.build());

    } catch (Exception e) {
      throw new LogProcessingException(e);
    }
  }
}
