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

import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogValidator;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.api.client.util.Preconditions;
import com.google.inject.Inject;
import lombok.AllArgsConstructor;

/** Implements validation process for the given {@link AuditLogRequest} */
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class ValidationProcessor implements LogValidator {

  /** Validates the given {@link AuditLogRequest} */
  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) throws IllegalArgumentException {
    Preconditions.checkArgument(auditLogRequest != null, "Input auditLogRequest is null");
    Preconditions.checkArgument(
        auditLogRequest.hasPayload(), "Input auditLogRequest does not have payload");
    Preconditions.checkArgument(
        auditLogRequest.getPayload().hasAuthenticationInfo(),
        "Input auditLogRequest does not have authentication info");
    Preconditions.checkArgument(
        auditLogRequest.getPayload().getResourceName() != null
            && !auditLogRequest.getPayload().getResourceName().isBlank(),
        "Input auditLogRequest does not have resource name");
    Preconditions.checkArgument(
        auditLogRequest.getPayload().getServiceName() != null
            && !auditLogRequest.getPayload().getServiceName().isBlank(),
        "Input auditLogRequest does not have service name");
    Preconditions.checkArgument(
        auditLogRequest.getPayload().getMethodName() != null
            && !auditLogRequest.getPayload().getMethodName().isBlank(),
        "Input auditLogRequest does not have method name");
    return auditLogRequest;
  }
}
