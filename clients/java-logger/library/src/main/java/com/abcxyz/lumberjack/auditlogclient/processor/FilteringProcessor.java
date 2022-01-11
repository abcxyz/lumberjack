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
import com.google.inject.Inject;
import com.google.inject.name.Named;
import java.util.List;
import java.util.regex.Pattern;
import lombok.AllArgsConstructor;
import lombok.Getter;

/** Implements filtering process for the given {@link AuditLogRequest} */
@AllArgsConstructor
@Getter
public class FilteringProcessor implements LogMutator {

  /**
   * includePatterns is list of regex pattern, When an audit log request has a principal email that
   * matches one the regular expressions, the audit log request is allowed for further processing
   */
  private final List<Pattern> includePatterns;
  /**
   * excludePatterns is list of regex pattern, When an audit log request has a principal email that
   * matches one the regular expressions, the audit log request is dropped for further processing
   */
  private final List<Pattern> excludePatterns;

  /**
   * Process uses include and exclude regex list to match with PrincipleEmail from the {@link
   * com.google.cloud.audit.AuthenticationInfo} to filter the {@link AuditLogRequest}
   */
  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) throws IllegalArgumentException {
    if (includePatterns.isEmpty() && excludePatterns.isEmpty()) {
      return auditLogRequest;
    }
    String principalEmail =
        auditLogRequest.getPayload().getAuthenticationInfo().getPrincipalEmail();
    if (includePatterns.stream().anyMatch(p -> p.matcher(principalEmail).matches())) {
      return auditLogRequest;
    }
    if (excludePatterns.isEmpty()) {
      throw new IllegalArgumentException(
          "request.Payload.AuthenticationInfo.PrincipalEmail not present in includePattern List");
    }

    if (excludePatterns.stream().anyMatch(p -> p.matcher(principalEmail).matches())) {
      throw new IllegalArgumentException(
          "request.Payload.AuthenticationInfo.PrincipalEmail " + "present in excludePattern List");
    }
    return auditLogRequest;
  }
}
