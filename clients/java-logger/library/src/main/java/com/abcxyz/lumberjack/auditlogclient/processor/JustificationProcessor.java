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

import com.abcxyz.jvs.JvsClient;
import com.abcxyz.lumberjack.auditlogclient.AuditLoggingServerInterceptor;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogMutator;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.auth0.jwk.JwkException;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.google.api.client.util.Preconditions;
import com.google.api.client.util.StringUtils;
import com.google.cloud.audit.AuditLog;
import com.google.inject.Inject;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.ListValue;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import com.google.protobuf.util.JsonFormat;
import lombok.AllArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.binary.Base64;

@Slf4j
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class JustificationProcessor implements LogMutator {
  private static final String JUSTIFICATION_LOG_METADATA_KEY = "justification";
  private final JvsClient jvs;

  /**
   * Extracts and validates the {@code jvsToken} and populates the {@link AuditLogRequest} with the
   * justification info.
   *
   * @param auditLogRequest Audit log request to be processed.
   * @return The {@link AuditLogRequest} with justification filled.
   * @throws LogProcessingException when it fails to populate the justification info.
   */
  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) throws LogProcessingException {
    Preconditions.checkArgument(
        auditLogRequest.hasPayload(), "audit log request doesn't have payload");

    AuditLogRequest.Builder auditLogRequestBuilder = auditLogRequest.toBuilder();
    AuditLog.Builder auditLogBuilder = auditLogRequest.getPayload().toBuilder();

    Value jvsToken =
        auditLogRequest
            .getContext()
            .getFieldsMap()
            .get(AuditLoggingServerInterceptor.JUSTIFICATION_TOKEN_HEADER_KEY);
    if (jvsToken == null || "".equals(jvsToken.getStringValue())) {
      log.info("no justification token found in AuditLogRequest");
      return auditLogRequest;
    }

    auditLogBuilder =
        this.auditLogBuilderWithJustification(jvsToken.getStringValue(), auditLogBuilder);

    return auditLogRequestBuilder.setPayload(auditLogBuilder.build()).build();
  }

  /**
   * Validates the given {@code jvsToken} and sets the justification info in the given {@code
   * auditLogBuilder}.
   *
   * @param jvsToken A JWT produced by JVS.
   * @param auditLogBuilder Audit log builder.
   * @return A new audit log builder with justification filled.
   * @throws LogProcessingException when it fails to set the justification info.
   */
  public AuditLog.Builder auditLogBuilderWithJustification(
      String jvsToken, AuditLog.Builder auditLogBuilder) throws LogProcessingException {

    AuditLog.Builder auditLogBuilderCopy = auditLogBuilder.build().toBuilder();
    try {
      DecodedJWT jwt = jvs.validateJWT(jvsToken);
      String jsonString = StringUtils.newStringUtf8(Base64.decodeBase64(jwt.getPayload()));
      Struct.Builder justificationStructBuilder = Struct.newBuilder();
      JsonFormat.parser().merge(jsonString, justificationStructBuilder);

      // Handle 'aud' claim properly.
      // Per https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.3, 'aud' is technically a
      // list.
      // But in the JWT json payload, often 'aud' is a single value which is acceptable.
      // This (a list vs. a single value) would cause a difference in the justification proto
      // struct, and it would cause downstream schema (e.g. BigQuery table) difference.
      if (justificationStructBuilder.containsFields("aud")) {
        Value audValue = justificationStructBuilder.getFieldsMap().get("aud");
        // If we find 'aud' is a single string value, convert it to be a list.
        if (audValue.hasStringValue()) {
          Value.Builder audBuilder =
              Value.newBuilder().setListValue(ListValue.newBuilder().addValues(audValue));
          justificationStructBuilder.putFields("aud", audBuilder.build());
        }
      }

      Struct justificationStruct = justificationStructBuilder.build();

      if (!auditLogBuilderCopy.hasMetadata()) {
        auditLogBuilderCopy.setMetadata(Struct.newBuilder().build());
      }

      Struct.Builder metadataBuilder = auditLogBuilderCopy.getMetadata().toBuilder();
      metadataBuilder.putFields(
          JUSTIFICATION_LOG_METADATA_KEY,
          Value.newBuilder().setStructValue(justificationStruct).build());
      auditLogBuilderCopy.setMetadata(metadataBuilder.build());

    } catch (JwkException | InvalidProtocolBufferException e) {
      throw new LogProcessingException(e);
    }

    return auditLogBuilderCopy;
  }
}
