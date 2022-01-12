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
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.times;
import static org.mockito.Mockito.verify;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.config.JwtSpecification;
import com.abcxyz.lumberjack.auditlogclient.config.SecurityContext;
import com.abcxyz.lumberjack.auditlogclient.config.Selector;
import com.google.cloud.audit.AuditLog;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import io.grpc.Metadata;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class AuditLoggingServerInterceptorTests {
  /*
   * { "sub": "1234567890",
   *   "name": "John Doe",
   *   "iat": 1516239022,
   *   "email": "me@example.com" }
   */
  private static final String ENCODED_Jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJlbWFpbCI6Im1lQGV4YW1wbGUuY29tIn0.6hBdfWsZcIn4crnRNBSMgztRaacHWmZmAtbaOc-efnI";

  @Mock
  LoggingClient loggingClient;
  @Mock
  AuditLoggingConfiguration auditLoggingConfiguration;
  AuditLoggingServerInterceptor interceptor;

  @BeforeEach
  public void setup() {
    interceptor = new AuditLoggingServerInterceptor(auditLoggingConfiguration, loggingClient);
  }

  @Test
  public void getsPrincipalFromJwt() {
    SecurityContext securityContext = mock(SecurityContext.class);
    doReturn(securityContext).when(auditLoggingConfiguration).getSecurityContext();
    List<JwtSpecification> specifications = new ArrayList<>();
    String key = "jwt-key";
    String prefix = "jwt-prefix ";
    specifications.add(new JwtSpecification(key, prefix, null));
    specifications.add(new JwtSpecification("not-found-key", "not-found", null));
    doReturn(specifications).when(securityContext).getJwtSpecifications();

    Metadata headers = new Metadata();
    Metadata.Key jwtKey =
        Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
    headers.put(jwtKey, prefix + ENCODED_Jwt);
    Metadata.Key otherKey =
        Metadata.Key.of("other-key", Metadata.ASCII_STRING_MARSHALLER);
    headers.put(otherKey, "irrelevant");

    Optional<String> returnVal = interceptor.getPrincipalFromJwt(headers);
    assertThat(returnVal.isPresent()).isTrue();
    assertThat(returnVal.get()).isEqualTo("me@example.com");
  }

  @Test
  public void getsPrincipalFromJwt_CaseInsensitive() {
    SecurityContext securityContext = mock(SecurityContext.class);
    doReturn(securityContext).when(auditLoggingConfiguration).getSecurityContext();
    List<JwtSpecification> specifications = new ArrayList<>();
    String key = "jwt-key";
    String prefix = "jwt-prefix ";
    specifications.add(new JwtSpecification(key, prefix, null));
    specifications.add(new JwtSpecification("not-found-key", "not-found", null));
    doReturn(specifications).when(securityContext).getJwtSpecifications();

    Metadata headers = new Metadata();
    Metadata.Key jwtKey =
        Metadata.Key.of(key.toUpperCase(), Metadata.ASCII_STRING_MARSHALLER);
    headers.put(jwtKey, prefix + ENCODED_Jwt);
    Metadata.Key otherKey =
        Metadata.Key.of("other-key", Metadata.ASCII_STRING_MARSHALLER);
    headers.put(otherKey, "irrelevant");

    Optional<String> returnVal = interceptor.getPrincipalFromJwt(headers);
    assertThat(returnVal.isPresent()).isTrue();
    assertThat(returnVal.get()).isEqualTo("me@example.com");
  }

  @Test
  public void getsPrincipalFromJwt_NoMatch() {
    SecurityContext securityContext = mock(SecurityContext.class);
    doReturn(securityContext).when(auditLoggingConfiguration).getSecurityContext();
    List<JwtSpecification> specifications = new ArrayList<>();
    String key = "jwt-key";
    String prefix = "jwt-prefix ";
    specifications.add(new JwtSpecification(key, prefix, null));
    specifications.add(new JwtSpecification("not-found-key", "not-found", null));
    doReturn(specifications).when(securityContext).getJwtSpecifications();

    Metadata headers = new Metadata();
    Metadata.Key otherKey =
        Metadata.Key.of("other-key", Metadata.ASCII_STRING_MARSHALLER);
    headers.put(otherKey, "irrelevant");

    Optional<String> returnVal = interceptor.getPrincipalFromJwt(headers);
    assertThat(returnVal.isPresent()).isFalse();
  }

  @Test
  public void convertsMessageToStruct() {
    AuditLog.Builder builder = AuditLog.newBuilder();
    builder.setServiceName("test-service");
    builder.setMethodName("test-method");
    Struct struct = interceptor.messageToStruct(builder.build());
    Struct.Builder structBuilder = Struct.newBuilder();
    structBuilder.putFields("serviceName", Value.newBuilder().setStringValue("test-service").build());
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
}
