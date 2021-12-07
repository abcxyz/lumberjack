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

package com.abcxyz.lumberjack.auditlogclient;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.verify;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.auditlogclient.processor.CloudLoggingProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.FilteringProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.auditlogclient.processor.RemoteProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.RuntimeInfoProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.ValidationProcessor;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class LoggingClientTests {

  @Mock
  ValidationProcessor validationProcessor;
  @Mock
  CloudLoggingProcessor cloudLoggingProcessor;
  @Mock
  RemoteProcessor remoteProcessor;
  @Mock
  FilteringProcessor filteringProcessor;
  @Mock
  RuntimeInfoProcessor runtimeInfoProcessor;

  @InjectMocks
  LoggingClientBuilder loggingClientBuilder;

  @Test
  void successfulClientCreate() {
    LoggingClient loggingClient = loggingClientBuilder.withDefaultProcessors().build();
    assertThat(loggingClient.getValidators().size()).isEqualTo(1);
    assertThat(loggingClient.getMutators().size()).isEqualTo(1);
    assertThat(loggingClient.getBackends().size()).isEqualTo(1);
  }

  @Test
  void multipleCallsWithSameProcessorMethodDoesNotAddProcessorMultipleTimes() {
    LoggingClient loggingClient = loggingClientBuilder
        .withValidationProcessor()
        .withValidationProcessor()
        .withFilteringProcessor()
        .withFilteringProcessor()
        .build();
    assertThat(loggingClient.getValidators().size()).isEqualTo(1);
    assertThat(loggingClient.getMutators().size()).isEqualTo(1);
    assertThat(loggingClient.getBackends().size()).isEqualTo(0);
  }

  @Test
  void logMethodCallsValidateProcessorTest() throws LogProcessingException {
    LoggingClient loggingClient = loggingClientBuilder.withValidationProcessor().build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    loggingClient.log(logRequest);
    verify(validationProcessor).process(logRequest);
  }

  @Test
  void logMethodCallsRemoteServiceLoggingProcessorTest() throws LogProcessingException {
    LoggingClient loggingClient = loggingClientBuilder.withRemoteProcessor().build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    loggingClient.log(logRequest);
    verify(remoteProcessor).process(logRequest);
  }

  @Test
  void logMethodCallsFilteringProcessorTest() throws LogProcessingException {
    LoggingClient loggingClient = loggingClientBuilder.withFilteringProcessor().build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    loggingClient.log(logRequest);
    verify(filteringProcessor).process(logRequest);
  }

  @Test
  void logMethodCallsRuntimeInfoProcessorTest() throws LogProcessingException {
    LoggingClient loggingClient = loggingClientBuilder.withRuntimeInfoProcessor().build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    loggingClient.log(logRequest);
    verify(runtimeInfoProcessor).process(logRequest);
  }

  @Test
  void logMethodCallsCloudLoggingProcessorTest() throws LogProcessingException {
    LoggingClient loggingClient = loggingClientBuilder.withCloudLoggingProcessor().build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    loggingClient.log(logRequest);
    verify(cloudLoggingProcessor).process(logRequest);
  }

  @Test
  void logCallsMultipleProcessorsWhenClientHasMultiple() throws LogProcessingException {
    LoggingClient loggingClient = loggingClientBuilder
        .withValidationProcessor()
        .withCloudLoggingProcessor()
        .withRemoteProcessor()
        .build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    doReturn(logRequest).when(validationProcessor).process(logRequest);
    doReturn(logRequest).when(cloudLoggingProcessor).process(logRequest);
    loggingClient.log(logRequest);
    verify(validationProcessor).process(logRequest);
    verify(cloudLoggingProcessor).process(logRequest);
    verify(remoteProcessor).process(logRequest);
  }

  @Test
  void logCallsMultipleProcessorsInCorrectOrderWhenClientHasMultiple()
      throws LogProcessingException {
    LoggingClient loggingClient = loggingClientBuilder
        .withValidationProcessor()
        .withCloudLoggingProcessor()
        .withRemoteProcessor()
        .withFilteringProcessor()
        .withRuntimeInfoProcessor()
        .build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    doReturn(logRequest).when(validationProcessor).process(logRequest);
    doReturn(logRequest).when(filteringProcessor).process(logRequest);
    doReturn(logRequest).when(runtimeInfoProcessor).process(logRequest);
    doReturn(logRequest).when(cloudLoggingProcessor).process(logRequest);
    loggingClient.log(logRequest);
    verify(validationProcessor).process(logRequest);
    verify(filteringProcessor).process(logRequest);
    verify(runtimeInfoProcessor).process(logRequest);
    verify(cloudLoggingProcessor).process(logRequest);
    verify(remoteProcessor).process(logRequest);
  }
}
