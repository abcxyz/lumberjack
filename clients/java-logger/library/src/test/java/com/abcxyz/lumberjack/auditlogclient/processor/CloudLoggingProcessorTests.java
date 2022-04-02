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
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.lenient;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogType;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.cloud.logging.LogEntry;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.Payload;
import com.google.cloud.logging.Synchronicity;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Set;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.ArgumentMatchers;
import org.mockito.Captor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class CloudLoggingProcessorTests {

  public static final String LOG_NAME_UNSPECIFIED =
      URLEncoder.encode("audit.abcxyz/unspecified", StandardCharsets.UTF_8);
  public static final String LOG_NAME_DATA_ACCESS =
      URLEncoder.encode("audit.abcxyz/data_access", StandardCharsets.UTF_8);

  @Spy private ObjectMapper mapper;
  @Mock private Logging logging;
  @Mock private AuditLoggingConfiguration auditLoggingConfiguration;
  @Captor private ArgumentCaptor<Set<LogEntry>> logEntryCaptor;
  @InjectMocks private CloudLoggingProcessor cloudLoggingProcessor;

  @BeforeEach
  void setup() {
    lenient().doReturn(LogMode.LOG_MODE_UNSPECIFIED).when(auditLoggingConfiguration).getLogMode();
  }

  @Test
  void shouldInvokeCloudLoggerWithLumberjackLogName() throws LogProcessingException {
    cloudLoggingProcessor.process(AuditLogRequest.getDefaultInstance());
    verify(logging).write(logEntryCaptor.capture());
    verify(logging).flush();
    LogEntry logEntry =
        logEntryCaptor.getValue().stream()
            .findFirst()
            .orElse(LogEntry.newBuilder(Payload.StringPayload.of("")).build());
    assertThat(logEntry.getLogName()).isEqualTo(LOG_NAME_UNSPECIFIED);
  }

  @Test
  void shouldSetCorrectLogNameGivenLogType() throws LogProcessingException {
    cloudLoggingProcessor.process(
        AuditLogRequest.getDefaultInstance()
            .newBuilderForType()
            .setType(LogType.DATA_ACCESS)
            .build());
    verify(logging).write(logEntryCaptor.capture());
    verify(logging).flush();
    LogEntry logEntry =
        logEntryCaptor.getValue().stream()
            .findFirst()
            .orElse(LogEntry.newBuilder(Payload.StringPayload.of("")).build());
    assertThat(logEntry.getLogName()).isEqualTo(LOG_NAME_DATA_ACCESS);
  }

  @Test
  void throwsExceptionWhenEncountersAuditLogProtoIssue() throws JsonProcessingException {
    doThrow(JsonProcessingException.class)
        .when(mapper)
        .readValue(anyString(), ArgumentMatchers.<TypeReference<Object>>any());
    assertThrows(
        LogProcessingException.class,
        () -> cloudLoggingProcessor.process(AuditLogRequest.getDefaultInstance()));
    verify(logging).flush();
  }

  @Test
  void setsSynchronicityToAsyncWhenLogModeIsNotFailClose() throws LogProcessingException {
    cloudLoggingProcessor.process(AuditLogRequest.getDefaultInstance());
    // By default, should be async.
    verify(logging, never()).setWriteSynchronicity(Synchronicity.SYNC);
    verify(logging).flush();
  }

  @Test
  void setsSynchronicityToSyncWhenLogModeIsFailClose() throws LogProcessingException {
    lenient().doReturn(LogMode.FAIL_CLOSE).when(auditLoggingConfiguration).getLogMode();
    cloudLoggingProcessor.process(AuditLogRequest.getDefaultInstance());
    // Fail close should make the calls sync.
    verify(logging).setWriteSynchronicity(Synchronicity.SYNC);
    verify(logging).flush();
  }
}
