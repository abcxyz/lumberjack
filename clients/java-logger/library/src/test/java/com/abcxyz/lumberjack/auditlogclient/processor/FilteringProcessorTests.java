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
import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import java.util.ArrayList;
import java.util.List;
import java.util.regex.Pattern;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class FilteringProcessorTests {

  AuthenticationInfo authenticationInfo =
      AuthenticationInfo.newBuilder().setPrincipalEmail("foo@google.com").build();
  AuditLog auditLog = AuditLog.newBuilder().setAuthenticationInfo(authenticationInfo).build();
  AuditLogRequest auditLogRequest = AuditLogRequest.newBuilder().setPayload(auditLog).build();

  @Test
  void withIncludeAndExcludeEmptyShouldPass() throws LogProcessingException {
    FilteringProcessor filteringProcessor =
        new FilteringProcessor(new ArrayList<>(), new ArrayList<>());
    assertThat(filteringProcessor.getExcludePatterns().isEmpty()).isTrue();
    assertThat(filteringProcessor.getIncludePatterns().isEmpty()).isTrue();
    assertDoesNotThrow(() -> filteringProcessor.process(auditLogRequest));
  }

  @Test
  void withIncludePresentExcludeEmptyShouldPass() throws LogProcessingException {
    List<Pattern> includes = new ArrayList<>(List.of(Pattern.compile("foo@google.com")));
    FilteringProcessor filteringProcessor = new FilteringProcessor(includes, new ArrayList<>());
    assertThat(filteringProcessor.getExcludePatterns().isEmpty()).isTrue();
    assertThat(filteringProcessor.getIncludePatterns().size()).isEqualTo(1);
    assertDoesNotThrow(() -> filteringProcessor.process(auditLogRequest));
  }

  @Test
  void withIncludeNotMatchedExcludeEmptyShouldFail() throws LogProcessingException {
    List<Pattern> includes = new ArrayList<>(List.of(Pattern.compile("@google.com")));
    FilteringProcessor filteringProcessor = new FilteringProcessor(includes, new ArrayList<>());
    assertThat(filteringProcessor.getExcludePatterns().isEmpty()).isTrue();
    assertThat(filteringProcessor.getIncludePatterns().size()).isEqualTo(1);
    assertThrows(
        PreconditionFailedException.class, () -> filteringProcessor.process(auditLogRequest));
  }

  @Test
  void withIncludeEmptyExcludeMatchedShouldFail() throws LogProcessingException {
    List<Pattern> excludes = new ArrayList<>(List.of(Pattern.compile("foo@google.com")));
    FilteringProcessor filteringProcessor = new FilteringProcessor(new ArrayList<>(), excludes);
    assertThat(filteringProcessor.getIncludePatterns().isEmpty()).isTrue();
    assertThat(filteringProcessor.getExcludePatterns().size()).isEqualTo(1);
    assertThrows(
        PreconditionFailedException.class, () -> filteringProcessor.process(auditLogRequest));
  }

  @Test
  void withIncludeEmptyExcludeNotMatchedShouldPass() throws LogProcessingException {
    List<Pattern> excludes = new ArrayList<>(List.of(Pattern.compile("@google.com")));
    FilteringProcessor filteringProcessor = new FilteringProcessor(new ArrayList<>(), excludes);
    assertThat(filteringProcessor.getIncludePatterns().isEmpty()).isTrue();
    assertThat(filteringProcessor.getExcludePatterns().size()).isEqualTo(1);
    assertDoesNotThrow(() -> filteringProcessor.process(auditLogRequest));
  }

  @Test
  void withIncludeMismatchedExcludeMatchedShouldFail() throws LogProcessingException {
    List<Pattern> includes = new ArrayList<>(List.of(Pattern.compile("@google.com")));
    List<Pattern> excludes = new ArrayList<>(List.of(Pattern.compile("foo@google.com")));
    FilteringProcessor filteringProcessor = new FilteringProcessor(includes, excludes);
    assertThat(filteringProcessor.getIncludePatterns().size()).isEqualTo(1);
    assertThat(filteringProcessor.getExcludePatterns().size()).isEqualTo(1);
    assertThrows(
        PreconditionFailedException.class, () -> filteringProcessor.process(auditLogRequest));
  }

  @Test
  void withIncludeMatchedExcludeMismatchedShouldPass() throws LogProcessingException {
    List<Pattern> includes = new ArrayList<>(List.of(Pattern.compile("foo@google.com")));
    List<Pattern> excludes = new ArrayList<>(List.of(Pattern.compile("@google.com")));
    FilteringProcessor filteringProcessor = new FilteringProcessor(includes, excludes);
    assertThat(filteringProcessor.getIncludePatterns().size()).isEqualTo(1);
    assertThat(filteringProcessor.getExcludePatterns().size()).isEqualTo(1);
    assertDoesNotThrow(() -> filteringProcessor.process(auditLogRequest));
  }

  @Test
  void withMultipleIncludeAndMultipleExcludeShouldPass() throws LogProcessingException {
    List<Pattern> includes =
        new ArrayList<>(
            List.of(Pattern.compile("foo@google.com"), Pattern.compile("bar@google.com")));
    List<Pattern> excludes =
        new ArrayList<>(
            List.of(Pattern.compile("@google.com"), Pattern.compile("bar1@google.com")));
    FilteringProcessor filteringProcessor = new FilteringProcessor(includes, excludes);
    assertThat(filteringProcessor.getIncludePatterns().size()).isEqualTo(2);
    assertThat(filteringProcessor.getExcludePatterns().size()).isEqualTo(2);
    assertDoesNotThrow(() -> filteringProcessor.process(auditLogRequest));
  }
}
