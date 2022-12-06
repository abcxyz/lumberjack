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
import static org.mockito.Mockito.verify;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.logging.LogEntry;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.Operation;
import com.google.cloud.logging.Payload;
import com.google.cloud.logging.Payload.JsonPayload;
import com.google.logging.v2.LogEntryOperation;
import com.google.protobuf.Timestamp;
import java.time.Instant;
import java.util.Map;
import java.util.Set;
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

  public static final String TEST_RESOURCE = "TEST_RESOURCE";
  public static final String OPERATION_ID = "OPERATION_ID";
  public static final String OPERATION_PRODUCER = "OPERATION_PRODUCER";
  public static final String LABEL_KEY = "FOO";
  public static final String LABEL_VALUE = "BAR";

  @Spy private ObjectMapper mapper;
  @Mock private Logging logging;
  @Mock private AuditLoggingConfiguration auditLoggingConfiguration;
  @Captor private ArgumentCaptor<Set<LogEntry>> logEntryCaptor;
  @InjectMocks private CloudLoggingProcessor cloudLoggingProcessor;

  @Test
  void shouldWriteCorrectLogEntry() throws LogProcessingException {
    AuditLog payload = AuditLog.newBuilder().setResourceName(TEST_RESOURCE).build();
    LogEntryOperation logEntryOperation =
        LogEntryOperation.newBuilder().setId(OPERATION_ID).setProducer(OPERATION_PRODUCER).build();
    Instant now = Instant.now();
    Timestamp timestamp =
        Timestamp.newBuilder().setSeconds(now.getEpochSecond()).setNanos(now.getNano()).build();
    cloudLoggingProcessor.process(
        AuditLogRequest.getDefaultInstance()
            .newBuilderForType()
            .setPayload(payload)
            .putLabels(LABEL_KEY, LABEL_VALUE)
            .setOperation(logEntryOperation)
            .setTimestamp(timestamp)
            .build());
    verify(logging).write(logEntryCaptor.capture());
    verify(logging).flush();
    LogEntry logEntry =
        logEntryCaptor.getValue().stream()
            .findFirst()
            .orElse(LogEntry.newBuilder(Payload.StringPayload.of("")).build());
    assertThat(((JsonPayload) logEntry.getPayload()).getDataAsMap())
        .isEqualTo(Map.of("resource_name", TEST_RESOURCE));
    assertThat(logEntry.getLabels()).isEqualTo(Map.of(LABEL_KEY, LABEL_VALUE));
    assertThat(logEntry.getOperation()).isEqualTo(Operation.of(OPERATION_ID, OPERATION_PRODUCER));
    assertThat(logEntry.getInstantTimestamp()).isEqualTo(now);
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
}
