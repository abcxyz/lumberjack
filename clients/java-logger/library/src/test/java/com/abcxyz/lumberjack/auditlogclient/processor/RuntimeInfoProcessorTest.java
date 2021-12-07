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

import static org.assertj.core.api.Assertions.assertThat;

import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import java.util.Optional;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class RuntimeInfoProcessorTest {

    Value monitoredResource = Value.newBuilder().setStructValue(Struct.newBuilder()
            .putFields("type", Value.newBuilder().setStringValue("gce_instance").build())
            .putFields("labels", Value.newBuilder().setStructValue(Struct.newBuilder()
                    .putFields("instanceId", Value.newBuilder().setStringValue("testID").build())
                    .putFields("zone", Value.newBuilder().setStringValue("testZone").build())
                    .build()).build())
            .build()).build();

    // auditLogRequest
    AuthenticationInfo authenticationInfo = AuthenticationInfo.newBuilder()
            .setPrincipalEmail("foo@google.com").build();
    AuditLog auditLog = AuditLog.newBuilder().setAuthenticationInfo(authenticationInfo).build();
    AuditLogRequest auditLogRequest = AuditLogRequest.newBuilder().setPayload(auditLog).build();

    @Test
    void shouldWriteMonitoredResourceToPayloadMetadata()
            throws LogProcessingException {
        RuntimeInfoProcessor runtimeInfoProcessor = new RuntimeInfoProcessor(
                Optional.of(monitoredResource));
        AuditLogRequest wantAuditLogRequest = runtimeInfoProcessor.process(auditLogRequest);
        assertThat(
                wantAuditLogRequest.getPayload().getMetadata().containsFields("originating_resource"))
                        .isTrue();
    }

    @Test
    void shouldAppendMonitoredResourcesToPayloadMetadata() {
        // add existing metadata
        AuditLogRequest.Builder auditLogRequestToUpdate = auditLogRequest.toBuilder();
        AuditLog.Builder auditLogToUpdate = auditLogRequest.getPayload().toBuilder();
        auditLogToUpdate.setMetadata(Struct.newBuilder()
                .putFields("existing_key", Value.newBuilder().setStringValue("existing_value").build())
                .build());
        auditLogRequest = auditLogRequestToUpdate.clearPayload().setPayload(auditLogToUpdate.build())
                .build();

        RuntimeInfoProcessor runtimeInfoProcessor = new RuntimeInfoProcessor(
                Optional.of(monitoredResource));
        AuditLogRequest wantAuditLogRequest = runtimeInfoProcessor.process(auditLogRequest);
        assertThat(wantAuditLogRequest.getPayload().getMetadata().getFieldsCount()).isEqualTo(2);
        assertThat(
                wantAuditLogRequest.getPayload().getMetadata().containsFields("originating_resource"))
                        .isTrue();
    }

    @Test
    void nullMonitoredResourceShouldLeaveMetadataUntouched() {
        RuntimeInfoProcessor runtimeInfoProcessor = new RuntimeInfoProcessor(Optional.empty());
        AuditLogRequest wantAuditLogRequest = runtimeInfoProcessor.process(auditLogRequest);
        assertThat(
                wantAuditLogRequest.getPayload().hasMetadata()).isFalse();
    }
}
