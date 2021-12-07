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

import static com.google.common.base.Preconditions.checkState;

import com.abcxyz.lumberjack.v1alpha1.AuditLogAgentGrpc;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

/**
 * Sends the {@link AuditLogRequest} to remote service via GRPC for log
 * processing.
 */
@Service
@RequiredArgsConstructor
public class RemoteProcessor implements LogProcessor {
  private final AuditLogAgentGrpc.AuditLogAgentBlockingStub blockingStub;

  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) {
    AuditLogResponse response = blockingStub.processLog(auditLogRequest);
    checkState(response != null && response.hasResult());

    AuditLogRequest.Builder requestToUpdate = auditLogRequest.toBuilder();
    AuditLogRequest result = response.getResult();
    requestToUpdate.clearLabels().putAllLabels(result.getLabelsMap());
    requestToUpdate.setPayload(result.getPayload());
    requestToUpdate.setType(result.getType());
    return requestToUpdate.build();
  }
}
