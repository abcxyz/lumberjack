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

package com.abcxyz.lumberjack.auditlogclient.config;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;

import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import io.grpc.Metadata;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class SecurityContextTest {
  /*
   * { "sub": "1234567890",
   *   "name": "John Doe",
   *   "iat": 1516239022,
   *   "email": "me@example.com" }
   */
  private static final String ENCODED_Jwt =
      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJlbWFpbCI6Im1lQGV4YW1wbGUuY29tIn0.6hBdfWsZcIn4crnRNBSMgztRaacHWmZmAtbaOc-efnI";
  private ObjectMapper mapper;

  @BeforeEach
  public void setupPropertyNaming() {
    mapper = new ObjectMapper(new YAMLFactory());
    mapper.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
  }

  @Test
  public void getsJwtSpecifications() {
    JwtSpecification jwtSpecification = mock(JwtSpecification.class);
    SecurityContext securityContext = new SecurityContext(List.of(jwtSpecification));
    assertThat(securityContext.getFromRawJwt().get(0)).isEqualTo(jwtSpecification);
  }

  @Test
  public void getsJwtSpecificationsDefaultIsEmpty() {
    SecurityContext securityContext = new SecurityContext();
    assertThat(securityContext.getSecuritySpecifications()).isEmpty();
  }

  @Test
  public void deserializesCorrectly() throws Exception {
    SecurityContext securityContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("jwt_context.yml"),
                AuditLoggingConfiguration.class)
            .getSecurityContext();
    SecurityContext expectedSecurityContext = new SecurityContext();
    JwtSpecification expectedJwtSpec = new JwtSpecification();
    expectedJwtSpec.setKey("x-jwt-assertion");
    expectedJwtSpec.setPrefix("Prefix ");
    JwksSpecification expectedJwksSpec = new JwksSpecification();
    expectedJwksSpec.setEndpoint("https://example.com");
    expectedJwksSpec.setObject("object");
    expectedJwtSpec.setJwks(expectedJwksSpec);
    expectedSecurityContext.setFromRawJwt(List.of(expectedJwtSpec));

    assertThat(securityContext).isEqualTo(expectedSecurityContext);
  }

  @Test
  public void deserializesCorrectly_default() throws Exception {
    SecurityContext securityContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("jwt_default.yml"),
                AuditLoggingConfiguration.class)
            .getSecurityContext();
    assertThat(securityContext.getFromRawJwt()).isEmpty();

    securityContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("jwt_default_2.yml"),
                AuditLoggingConfiguration.class)
            .getSecurityContext();
    assertThat(securityContext.getFromRawJwt()).isEmpty();
  }

  @Test
  public void failsWithNoSecurityContext() throws IOException {
    Assertions.assertThrows(
        JsonMappingException.class,
        () ->
            mapper.readValue(
                getClass().getClassLoader().getResourceAsStream("no_security_context.yml"),
                AuditLoggingConfiguration.class));
  }

  @Test
  public void getsPrincipalFromJwt() throws Exception {
    List<JwtSpecification> specifications = new ArrayList<>();
    String key = "jwt-key";
    String prefix = "jwt-prefix ";
    specifications.add(new JwtSpecification(key, prefix, null));
    specifications.add(new JwtSpecification("not-found-key", "not-found", null));
    SecurityContext securityContext = new SecurityContext(specifications);

    Metadata headers = new Metadata();
    Metadata.Key<String> jwtKey = Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
    headers.put(jwtKey, prefix + ENCODED_Jwt);
    Metadata.Key<String> otherKey = Metadata.Key.of("other-key", Metadata.ASCII_STRING_MARSHALLER);
    headers.put(otherKey, "irrelevant");

    Optional<String> returnVal = securityContext.getPrincipal(headers);
    assertThat(returnVal.isPresent()).isTrue();
    assertThat(returnVal.get()).isEqualTo("me@example.com");
  }

  @Test
  public void getsPrincipalFromJwt_NullPrefix() throws Exception {
    List<JwtSpecification> specifications = new ArrayList<>();
    String key = "jwt-key";
    specifications.add(new JwtSpecification(key, null, null));
    specifications.add(new JwtSpecification("not-found-key", "not-found", null));
    SecurityContext securityContext = new SecurityContext(specifications);

    Metadata headers = new Metadata();
    Metadata.Key<String> jwtKey = Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
    headers.put(jwtKey, ENCODED_Jwt);
    Metadata.Key<String> otherKey = Metadata.Key.of("other-key", Metadata.ASCII_STRING_MARSHALLER);
    headers.put(otherKey, "irrelevant");

    Optional<String> returnVal = securityContext.getPrincipal(headers);
    assertThat(returnVal.isPresent()).isTrue();
    assertThat(returnVal.get()).isEqualTo("me@example.com");
  }

  @Test
  public void getsPrincipalFromJwt_CaseInsensitive() throws Exception {
    List<JwtSpecification> specifications = new ArrayList<>();
    String key = "jwt-key";
    String prefix = "jwt-prefix ";
    specifications.add(new JwtSpecification(key, prefix, null));
    specifications.add(new JwtSpecification("not-found-key", "not-found", null));
    SecurityContext securityContext = new SecurityContext(specifications);

    Metadata headers = new Metadata();
    Metadata.Key<String> jwtKey =
        Metadata.Key.of(key.toUpperCase(), Metadata.ASCII_STRING_MARSHALLER);
    headers.put(jwtKey, prefix.toUpperCase() + ENCODED_Jwt);
    Metadata.Key<String> otherKey = Metadata.Key.of("other-key", Metadata.ASCII_STRING_MARSHALLER);
    headers.put(otherKey, "irrelevant");

    Optional<String> returnVal = securityContext.getPrincipal(headers);
    assertThat(returnVal.isPresent()).isTrue();
    assertThat(returnVal.get()).isEqualTo("me@example.com");
  }

  @Test
  public void getsPrincipalFromJwt_NoMatch() throws Exception {
    List<JwtSpecification> specifications = new ArrayList<>();
    String key = "jwt-key";
    String prefix = "jwt-prefix ";
    specifications.add(new JwtSpecification(key, prefix, null));
    specifications.add(new JwtSpecification("not-found-key", "not-found", null));
    SecurityContext securityContext = new SecurityContext(specifications);

    Metadata headers = new Metadata();
    Metadata.Key<String> otherKey = Metadata.Key.of("other-key", Metadata.ASCII_STRING_MARSHALLER);
    headers.put(otherKey, "irrelevant");

    Optional<String> returnVal = securityContext.getPrincipal(headers);
    assertThat(returnVal.isPresent()).isFalse();
  }
}
