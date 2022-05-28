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
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.lenient;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.config.BackendContext;
import com.abcxyz.lumberjack.auditlogclient.processor.CloudLoggingProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.FilteringProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LabelProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LocalLogProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.auditlogclient.processor.RemoteProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.RuntimeInfoProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.ValidationProcessor;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;
import com.google.inject.Injector;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class LoggingClientTests {
  @Mock ValidationProcessor validationProcessor;
  @Mock CloudLoggingProcessor cloudLoggingProcessor;
  @Mock RemoteProcessor remoteProcessor;
  @Mock LocalLogProcessor localLogProcessor;
  @Mock FilteringProcessor filteringProcessor;
  @Mock RuntimeInfoProcessor runtimeInfoProcessor;
  @Mock AuditLoggingConfiguration auditLoggingConfiguration;
  @Mock LabelProcessor labelProcessor;
  @Mock BackendContext backendContext;
  @Mock Injector injector;

  @InjectMocks LoggingClientBuilder loggingClientBuilder;

  @BeforeEach
  void setup() {
    lenient().doReturn(LogMode.LOG_MODE_UNSPECIFIED).when(auditLoggingConfiguration).getLogMode();

    // Ensure backend context set to log to remote
    lenient().doReturn(false).when(backendContext).localLoggingEnabled();
    lenient().doReturn(false).when(backendContext).cloudLoggingEnabled();
    lenient().doReturn(true).when(backendContext).remoteEnabled();
    lenient().doReturn(backendContext).when(auditLoggingConfiguration).getBackend();

    lenient().doReturn(validationProcessor).when(injector).getInstance(ValidationProcessor.class);
    lenient()
        .doReturn(cloudLoggingProcessor)
        .when(injector)
        .getInstance(CloudLoggingProcessor.class);
    lenient().doReturn(remoteProcessor).when(injector).getInstance(RemoteProcessor.class);
    lenient().doReturn(localLogProcessor).when(injector).getInstance(LocalLogProcessor.class);
    lenient().doReturn(filteringProcessor).when(injector).getInstance(FilteringProcessor.class);
    lenient().doReturn(runtimeInfoProcessor).when(injector).getInstance(RuntimeInfoProcessor.class);
    lenient().doReturn(labelProcessor).when(injector).getInstance(LabelProcessor.class);

    injector.injectMembers(this);
  }

  @Test
  void successfulClientCreate() {
    LoggingClient loggingClient = loggingClientBuilder.withDefaultProcessors().build();
    assertThat(loggingClient.getValidators().size()).isEqualTo(1);
    assertThat(loggingClient.getMutators().size()).isEqualTo(3);
    assertThat(loggingClient.getBackends().size()).isEqualTo(1);

    // We want filtering to occur before other mutators
    assertThat(loggingClient.getMutators().get(0).equals(filteringProcessor));
    assertThat(loggingClient.getMutators().get(1).equals(runtimeInfoProcessor));

    // If local is disabled and remote is enabled, only backend processor should be remote.
    assertThat(loggingClient.getBackends().get(0).equals(remoteProcessor));
  }

  @Test
  void successfulClientCreate_LocalBackend() {
    // Ensure backend context set to log to local
    doReturn(true).when(backendContext).localLoggingEnabled();
    doReturn(false).when(backendContext).remoteEnabled();

    LoggingClient loggingClient = loggingClientBuilder.withDefaultProcessors().build();
    assertThat(loggingClient.getValidators().size()).isEqualTo(1);
    assertThat(loggingClient.getMutators().size()).isEqualTo(3);
    assertThat(loggingClient.getBackends().size()).isEqualTo(1);

    // We want filtering to occur before other mutators
    assertThat(loggingClient.getMutators().get(0).equals(filteringProcessor));
    assertThat(loggingClient.getMutators().get(1).equals(runtimeInfoProcessor));

    // If remote is disabled and local is enabled, only backend processor should be local.
    assertThat(loggingClient.getBackends().get(0).equals(localLogProcessor));
  }

  @Test
  void successfulClientCreate_CloudLoggingBackend() {
    // Ensure backend context set to log to local
    doReturn(false).when(backendContext).localLoggingEnabled();
    doReturn(false).when(backendContext).remoteEnabled();
    doReturn(true).when(backendContext).cloudLoggingEnabled();

    LoggingClient loggingClient = loggingClientBuilder.withDefaultProcessors().build();
    assertThat(loggingClient.getValidators().size()).isEqualTo(1);
    assertThat(loggingClient.getMutators().size()).isEqualTo(3);
    assertThat(loggingClient.getBackends().size()).isEqualTo(1);

    // We want filtering to occur before other mutators
    assertThat(loggingClient.getMutators().get(0).equals(filteringProcessor));
    assertThat(loggingClient.getMutators().get(1).equals(runtimeInfoProcessor));

    // If remote is disabled and local is enabled, only backend processor should be local.
    assertThat(loggingClient.getBackends().get(0).equals(localLogProcessor));
  }

  @Test
  void successfulClientCreate_LocalAndRemoteBackends() {
    // Ensure backend context set to log to local
    doReturn(true).when(backendContext).localLoggingEnabled();
    doReturn(true).when(backendContext).remoteEnabled();

    LoggingClient loggingClient = loggingClientBuilder.withDefaultProcessors().build();
    assertThat(loggingClient.getValidators().size()).isEqualTo(1);
    assertThat(loggingClient.getMutators().size()).isEqualTo(3);

    // Both backends should be added
    assertThat(loggingClient.getBackends().size()).isEqualTo(2);

    // We want filtering to occur before other mutators
    assertThat(loggingClient.getMutators().get(0).equals(filteringProcessor));
    assertThat(loggingClient.getMutators().get(1).equals(runtimeInfoProcessor));
  }

  @Test
  void multipleCallsWithSameProcessorMethodDoesNotAddProcessorMultipleTimes() {
    LoggingClient loggingClient =
        loggingClientBuilder
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
    LoggingClient loggingClient =
        loggingClientBuilder
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
    LoggingClient loggingClient =
        loggingClientBuilder
            .withValidationProcessor()
            .withFilteringProcessor()
            .withRuntimeInfoProcessor()
            .withCloudLoggingProcessor()
            .withRemoteProcessor()
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

  @Test
  void failsOpen() throws LogProcessingException {
    // Set config to fail open
    doReturn(LogMode.BEST_EFFORT).when(auditLoggingConfiguration).getLogMode();
    LoggingClient loggingClient =
        loggingClientBuilder
            .withValidationProcessor()
            .withFilteringProcessor()
            .withRuntimeInfoProcessor()
            .withCloudLoggingProcessor()
            .withRemoteProcessor()
            .build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    when(validationProcessor.process(any())).thenThrow(new IllegalArgumentException());

    // No exception is thrown
    Assertions.assertDoesNotThrow(() -> loggingClient.log(logRequest));

    // No other processors were run after the exception was swallowed.
    verify(filteringProcessor, times(0)).process(logRequest);
    verify(runtimeInfoProcessor, times(0)).process(logRequest);
    verify(cloudLoggingProcessor, times(0)).process(logRequest);
    verify(remoteProcessor, times(0)).process(logRequest);
  }

  @Test
  void failsClose() throws LogProcessingException {
    // Set config to fail close
    doReturn(LogMode.FAIL_CLOSE).when(auditLoggingConfiguration).getLogMode();
    LoggingClient loggingClient =
        loggingClientBuilder
            .withValidationProcessor()
            .withFilteringProcessor()
            .withRuntimeInfoProcessor()
            .withCloudLoggingProcessor()
            .withRemoteProcessor()
            .build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().getDefaultInstanceForType();
    when(validationProcessor.process(logRequest)).thenThrow(new IllegalArgumentException());

    // Exception is thrown
    Assertions.assertThrows(LogProcessingException.class, () -> loggingClient.log(logRequest));

    // No other processors were run after the exception was thrown.
    verify(filteringProcessor, times(0)).process(logRequest);
    verify(runtimeInfoProcessor, times(0)).process(logRequest);
    verify(cloudLoggingProcessor, times(0)).process(logRequest);
    verify(remoteProcessor, times(0)).process(logRequest);
  }

  @Test
  void requestLogModeValueOverridesConfig() throws LogProcessingException {
    // Set config to fail open. We use lenient() here, as we don't expect this value to be used.
    lenient().doReturn(LogMode.BEST_EFFORT).when(auditLoggingConfiguration).getLogMode();
    LoggingClient loggingClient =
        loggingClientBuilder
            .withValidationProcessor()
            .withFilteringProcessor()
            .withRuntimeInfoProcessor()
            .withCloudLoggingProcessor()
            .withRemoteProcessor()
            .build();
    AuditLogRequest logRequest = AuditLogRequest.newBuilder().setMode(LogMode.FAIL_CLOSE).build();
    when(validationProcessor.process(logRequest)).thenThrow(new IllegalArgumentException());

    // Exception is thrown
    Assertions.assertThrows(LogProcessingException.class, () -> loggingClient.log(logRequest));

    // No other processors were run after the exception was thrown.
    verify(filteringProcessor, times(0)).process(logRequest);
    verify(runtimeInfoProcessor, times(0)).process(logRequest);
    verify(cloudLoggingProcessor, times(0)).process(logRequest);
    verify(remoteProcessor, times(0)).process(logRequest);
  }

  @Test
  void onlyUsedProcessorIsInjected() {
    LoggingClient loggingClient = loggingClientBuilder.withCloudLoggingProcessor().build();
    verify(injector).getInstance(CloudLoggingProcessor.class);
    verify(injector, times(0)).getInstance(RemoteProcessor.class);
    assertThat(loggingClient.getBackends().size()).isEqualTo(1);
    assertThat(loggingClient.getBackends().get(0).equals(cloudLoggingProcessor));
  }

  @Test
  void multipleCallsOnlyInjectsOneUniqueProcessor() {
    LoggingClient loggingClient =
        loggingClientBuilder.withCloudLoggingProcessor().withCloudLoggingProcessor().build();
    verify(injector, times(1)).getInstance(CloudLoggingProcessor.class);
    assertThat(loggingClient.getBackends().size()).isEqualTo(1);
    assertThat(loggingClient.getBackends().get(0).equals(cloudLoggingProcessor));
  }
}
