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
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.config.Selector;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.cloud.audit.AuditLog;
import com.google.logging.v2.LogEntryOperation;
import com.google.protobuf.Struct;
import com.google.protobuf.Timestamp;
import com.google.protobuf.Value;
import com.google.rpc.Code;
import com.google.rpc.Status;
import io.grpc.Metadata;
import io.grpc.StatusRuntimeException;
import java.time.Clock;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class AuditLoggingServerInterceptorTests {

  @Mock LoggingClient loggingClient;

  @Mock AuditLoggingConfiguration auditLoggingConfiguration;

  @Mock Clock clock;

  @InjectMocks AuditLoggingServerInterceptor interceptor;

  @Captor private ArgumentCaptor<AuditLogRequest> auditLogRequestCaptor;

  @Test
  public void convertsMessageToStruct() {
    AuditLog.Builder builder = AuditLog.newBuilder();
    builder.setServiceName("test-service");
    builder.setMethodName("test-method");
    Struct struct = interceptor.messageToStruct(builder.build());
    Struct.Builder structBuilder = Struct.newBuilder();
    structBuilder.putFields(
        "serviceName", Value.newBuilder().setStringValue("test-service").build());
    structBuilder.putFields("methodName", Value.newBuilder().setStringValue("test-method").build());
    assertThat(struct).isEqualTo(structBuilder.build());
  }

  @Test
  public void convertsMessageToStruct_Empty() {
    AuditLog.Builder builder = AuditLog.newBuilder();
    Struct struct = interceptor.messageToStruct(builder.build());
    Struct.Builder structBuilder = Struct.newBuilder();
    assertThat(struct).isEqualTo(structBuilder.build());
  }

  @Test
  public void convertsMessagesToStruct() {
    AuditLog.Builder builder = AuditLog.newBuilder();
    builder.setServiceName("test-service");
    builder.setMethodName("test-method");

    AuditLog.Builder builder2 = AuditLog.newBuilder();
    builder2.setServiceName("test-service-2");
    builder2.setMethodName("test-method-2");

    List<AuditLog> messages = List.of(builder.build(), builder2.build());
    Struct actual = interceptor.messagesToStruct(messages);

    Struct.Builder structBuilder = Struct.newBuilder();
    structBuilder.putFields(
        "request_list",
        Value.newBuilder()
            .setStringValue(
                "[service_name: \"test-service\"method_name: \"test-method\", service_name:"
                    + " \"test-service-2\"method_name: \"test-method-2\"]")
            .build());
    assertThat(actual).isEqualTo(structBuilder.build());
  }

  @Test
  public void convertsMessagesToStruct_Empty() {
    Struct actual = interceptor.messagesToStruct(Collections.emptyList());

    Struct.Builder structBuilder = Struct.newBuilder();
    structBuilder.putFields("request_list", Value.newBuilder().setStringValue("[]").build());
    assertThat(actual).isEqualTo(structBuilder.build());
  }

  @Test
  public void getsRelevantSelector() {
    List<Selector> selectors = new ArrayList<>();
    Selector selector1 = new Selector("*", null, null);
    selectors.add(selector1);
    Selector selector2 = new Selector("com.example.a", null, null);
    selectors.add(selector2);
    Selector selector3 = new Selector("com.example.a.*", null, null);
    selectors.add(selector3);
    Selector selector4 = new Selector("com.example.a.stuff", null, null);
    selectors.add(selector4);

    // We expect that no selector will be returned if none are available.
    doReturn(Collections.emptyList()).when(auditLoggingConfiguration).getRules();
    Optional<Selector> chosenSelector = interceptor.getRelevantSelector("other.package");
    assertThat(chosenSelector.isPresent()).isFalse();
    // We expect the cache to have nothing, therefore rules are read.
    verify(auditLoggingConfiguration, times(1)).getRules();

    // We expect that given there are relevant selectors, the correct one will be chosen.
    doReturn(selectors).when(auditLoggingConfiguration).getRules();
    chosenSelector = interceptor.getRelevantSelector("com.example.a.other-stuff");
    assertThat(chosenSelector.isPresent()).isTrue();
    assertThat(chosenSelector.get()).isEqualTo(selector3);
    // We expect the cache to not have this method, and therefore rules are read again.
    verify(auditLoggingConfiguration, times(2)).getRules();

    chosenSelector = interceptor.getRelevantSelector("com.example.a.other-stuff");
    assertThat(chosenSelector.isPresent()).isTrue();
    assertThat(chosenSelector.get()).isEqualTo(selector3);
    // We expect the cache to have this method, and therefore rules should not be read again.
    verify(auditLoggingConfiguration, times(2)).getRules();
  }

  @Test
  public void logsError() throws LogProcessingException {
    Selector selector = new Selector("*", null, null);
    AuditLog.Builder builder = AuditLog.newBuilder();
    doReturn(Clock.systemUTC().instant()).when(clock).instant();
    interceptor.logError(
        selector,
        Struct.getDefaultInstance(),
        null,
        new RuntimeException("Test Message"),
        builder,
        LogEntryOperation.getDefaultInstance());
    AuditLog.Builder expectedBuilder =
        AuditLog.newBuilder()
            .setStatus(
                Status.newBuilder()
                    .setMessage(Code.INTERNAL.name())
                    .setCode(Code.INTERNAL.getNumber())
                    .build());
    verify(loggingClient).log(auditLogRequestCaptor.capture());
    assertThat(auditLogRequestCaptor.getValue().getPayload().getStatus())
        .isEqualTo(expectedBuilder.getStatus())
        .usingRecursiveComparison();
  }

  @Test
  public void logsError_GRPC_Code() throws LogProcessingException {
    Selector selector = new Selector("*", null, null);
    AuditLog.Builder builder = AuditLog.newBuilder();
    doReturn(Clock.systemUTC().instant()).when(clock).instant();
    interceptor.logError(
        selector,
        Struct.getDefaultInstance(),
        null,
        new StatusRuntimeException(io.grpc.Status.FAILED_PRECONDITION),
        builder,
        LogEntryOperation.getDefaultInstance());
    AuditLog.Builder expectedBuilder =
        AuditLog.newBuilder()
            .setStatus(
                Status.newBuilder()
                    .setMessage(Code.FAILED_PRECONDITION.name())
                    .setCode(Code.FAILED_PRECONDITION.getNumber())
                    .build());
    verify(loggingClient).log(auditLogRequestCaptor.capture());
    assertThat(auditLogRequestCaptor.getValue().getPayload().getStatus())
        .isEqualTo(expectedBuilder.getStatus())
        .usingRecursiveComparison();
  }

  @Test
  public void logsError_CapturesTimestamp() throws LogProcessingException {
    Selector selector = new Selector("*", null, null);
    AuditLog.Builder builder = AuditLog.newBuilder();
    Instant now = Instant.parse("2022-04-18T22:11:22.333333Z");
    doReturn(now).when(clock).instant();
    interceptor.logError(
        selector,
        Struct.getDefaultInstance(),
        null,
        new StatusRuntimeException(io.grpc.Status.FAILED_PRECONDITION),
        builder,
        LogEntryOperation.getDefaultInstance());
    verify(loggingClient).log(auditLogRequestCaptor.capture());
    assertThat(auditLogRequestCaptor.getValue().hasTimestamp()).isTrue();
    assertThat(
        auditLogRequestCaptor
            .getValue()
            .getTimestamp()
            .equals(
                Timestamp.newBuilder()
                    .setSeconds(now.getEpochSecond())
                    .setNanos(now.getNano())
                    .build()));
  }

  @Test
  public void getAuditLogRequestContext() throws Exception {
    // Create Metadata
    Metadata md = new Metadata();
    Metadata.Key<String> metadataKey =
        Metadata.Key.of("justification-token", Metadata.ASCII_STRING_MARSHALLER);
    md.put(metadataKey, "token");

    Struct context = interceptor.getAuditLogRequestContext(md);
    assertEquals("token", context.getFieldsMap().get("justification-token").getStringValue());
  }

  @Test
  public void getAuditLogRequestContext_Empty() throws Exception {
    Struct context = interceptor.getAuditLogRequestContext(new Metadata());
    assertNull(context.getFieldsMap().get("justification-token"));
  }
}
