/*
 * Copyright 2022 Lumberjack authors (see AUTHORS file)
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

import static org.junit.Assert.assertThrows;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;

import com.abcxyz.jvs.JvsClient;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.auth0.jwk.JwkException;
import com.auth0.jwt.JWT;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.google.protobuf.ListValue;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import io.jsonwebtoken.Jwts;
import java.util.HashMap;
import java.util.Map;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class JustificationProcessorTests {

  @Mock JvsClient jvsClient;

  // auditLogRequest
  AuthenticationInfo authenticationInfo =
      AuthenticationInfo.newBuilder().setPrincipalEmail("user@example.com").build();
  AuditLog auditLog = AuditLog.newBuilder().setAuthenticationInfo(authenticationInfo).build();
  AuditLogRequest auditLogRequest = AuditLogRequest.newBuilder().setPayload(auditLog).build();

  @Test
  public void processShouldPopulateJustificationMetadata() throws Exception {
    // Create JWT
    Map<String, Object> claims = new HashMap<>();
    claims.put("id", "jwt-id");
    claims.put("role", "user");
    String token = Jwts.builder().setClaims(claims).setAudience("test-aud").compact();

    // Set up JVS mock to return the correct token
    DecodedJWT jwt = JWT.decode(token);
    doReturn(jwt).when(jvsClient).validateJWT(token);

    AuditLogRequest wantAuditLogReq = AuditLogRequest.parseFrom(auditLogRequest.toByteArray());
    Struct wantJustification =
        Struct.newBuilder()
            .putFields("id", Value.newBuilder().setStringValue("jwt-id").build())
            .putFields("role", Value.newBuilder().setStringValue("user").build())
            .putFields(
                "aud",
                Value.newBuilder()
                    .setListValue(
                        ListValue.newBuilder()
                            .addValues(Value.newBuilder().setStringValue("test-aud")))
                    .build())
            .build();
    Struct wantMetadata =
        Struct.newBuilder()
            .putFields(
                "justification", Value.newBuilder().setStructValue(wantJustification).build())
            .build();
    AuditLog wantAuditLog =
        wantAuditLogReq.getPayload().toBuilder().setMetadata(wantMetadata).build();
    wantAuditLogReq = wantAuditLogReq.toBuilder().setPayload(wantAuditLog).build();

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    AuditLogRequest gotAuditLogReq = processor.process(token, auditLogRequest);

    assertEquals(wantAuditLogReq, gotAuditLogReq);
  }

  @Test
  public void processShouldThrowExceptionWhenFailingToValidateToken() throws Exception {
    // Create JWT
    Map<String, Object> claims = new HashMap<>();
    claims.put("id", "jwt-id");
    claims.put("role", "user");
    String token = Jwts.builder().setClaims(claims).compact();

    // Set up JVS mock to throw exception
    doThrow(new JwkException("")).when(jvsClient).validateJWT(token);

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertThrows(LogProcessingException.class, () -> processor.process(token, auditLogRequest));
  }

  @Test
  public void processShouldThrowExceptionWithoutLogReqPayload() throws Exception {
    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertThrows(
        IllegalArgumentException.class,
        () -> processor.process("token", AuditLogRequest.newBuilder().build()));
  }

  @Test
  public void setLogJustificationShouldPopulateJustificationMetadata() throws Exception {
    // Create JWT
    Map<String, Object> claims = new HashMap<>();
    claims.put("id", "jwt-id");
    claims.put("role", "user");
    String token = Jwts.builder().setClaims(claims).setAudience("test-aud").compact();

    // Set up JVS mock to return the correct token
    DecodedJWT jwt = JWT.decode(token);
    doReturn(jwt).when(jvsClient).validateJWT(token);

    Struct wantJustification =
        Struct.newBuilder()
            .putFields("id", Value.newBuilder().setStringValue("jwt-id").build())
            .putFields("role", Value.newBuilder().setStringValue("user").build())
            .putFields(
                "aud",
                Value.newBuilder()
                    .setListValue(
                        ListValue.newBuilder()
                            .addValues(Value.newBuilder().setStringValue("test-aud")))
                    .build())
            .build();
    Struct wantMetadata =
        Struct.newBuilder()
            .putFields(
                "justification", Value.newBuilder().setStructValue(wantJustification).build())
            .build();
    AuditLog wantAuditLog = auditLog.toBuilder().setMetadata(wantMetadata).build();
    AuditLog.Builder gotAuditLogBuilder = auditLog.toBuilder();

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    gotAuditLogBuilder = processor.auditLogBuilderWithJustification(token, gotAuditLogBuilder);

    assertEquals(wantAuditLog, gotAuditLogBuilder.build());
  }

  @Test
  public void setLogJustificationShouldThrowExceptionWhenFailingToValidateToken() throws Exception {
    // Create JWT
    Map<String, Object> claims = new HashMap<>();
    claims.put("id", "jwt-id");
    claims.put("role", "user");
    String token = Jwts.builder().setClaims(claims).compact();

    // Set up JVS mock to throw exception
    doThrow(new JwkException("")).when(jvsClient).validateJWT(token);

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertThrows(
        LogProcessingException.class,
        () -> processor.auditLogBuilderWithJustification(token, auditLog.toBuilder()));
  }
}
