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

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.verifyNoInteractions;

import com.abcxyz.jvs.JvsClient;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.auth0.jwk.JwkException;
import com.auth0.jwt.JWT;
import com.auth0.jwt.interfaces.DecodedJWT;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.google.cloud.audit.RequestMetadata;
import com.google.protobuf.ListValue;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import com.google.rpc.context.AttributeContext.Request;
import io.jsonwebtoken.Jwts;
import java.util.HashMap;
import java.util.List;
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

    Map<String, String> justification = new HashMap<>();
    justification.put("category", "explanation");
    justification.put("value", "need-access");
    claims.put("justs", List.of(justification));

    String token = Jwts.builder().setClaims(claims).setAudience("test-aud").compact();
    AuditLogRequest auditLogReqWithToken =
        auditLogRequest.toBuilder().setJustificationToken(token).setPayload(auditLog).build();

    // Set up JVS mock to return the correct token
    DecodedJWT jwt = JWT.decode(token);
    doReturn(jwt).when(jvsClient).validateJWT(token, "user@example.com");

    AuditLogRequest wantAuditLogReq = AuditLogRequest.parseFrom(auditLogReqWithToken.toByteArray());
    Struct wantJustificationsClaim =
        Struct.newBuilder()
            .putFields("category", Value.newBuilder().setStringValue("explanation").build())
            .putFields("value", Value.newBuilder().setStringValue("need-access").build())
            .build();
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
            .putFields(
                "justs",
                Value.newBuilder()
                    .setListValue(
                        ListValue.newBuilder()
                            .addValues(Value.newBuilder().setStructValue(wantJustificationsClaim)))
                    .build())
            .build();
    Struct wantMetadata =
        Struct.newBuilder()
            .putFields(
                "justification", Value.newBuilder().setStructValue(wantJustification).build())
            .build();
    RequestMetadata wantRequestMetadata =
        RequestMetadata.newBuilder()
            .setRequestAttributes(
                Request.newBuilder()
                    .setReason("[{\"category\":\"explanation\",\"value\":\"need-access\"}]")
                    .build())
            .build();
    AuditLog wantAuditLog =
        wantAuditLogReq.getPayload().toBuilder()
            .setMetadata(wantMetadata)
            .setRequestMetadata(wantRequestMetadata)
            .build();
    wantAuditLogReq = wantAuditLogReq.toBuilder().setPayload(wantAuditLog).build();

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    AuditLogRequest gotAuditLogReq = processor.process(auditLogReqWithToken);

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
    doThrow(new JwkException("")).when(jvsClient).validateJWT(token, "user@example.com");

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertThrows(
        LogProcessingException.class,
        () -> processor.process(auditLogRequest.toBuilder().setJustificationToken(token).build()));
  }

  @Test
  public void processShouldThrowExceptionWithoutLogReqPayload() throws Exception {
    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertThrows(
        IllegalArgumentException.class,
        () -> processor.process(AuditLogRequest.newBuilder().build()));
  }

  @Test
  public void processShouldThrowExceptionWithoutJVSToken() throws Exception {
    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertThrows(LogProcessingException.class, () -> processor.process(auditLogRequest));
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
    doReturn(jwt).when(jvsClient).validateJWT(token, "user@example.com");

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
    doThrow(new JwkException("")).when(jvsClient).validateJWT(token, "user@example.com");

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertThrows(
        LogProcessingException.class,
        () -> processor.auditLogBuilderWithJustification(token, auditLog.toBuilder()));
  }

  @Test
  public void getJustificationList() {
    Map<String, String> justification = new HashMap<>();
    justification.put("category", "explanation");
    justification.put("value", "need-access");
    Map<String, Object> claims = Map.of("justs", List.of(justification));

    String token = Jwts.builder().setClaims(claims).compact();
    DecodedJWT jwt = JWT.decode(token);

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertEquals(List.of(justification), processor.getJustificationList(jwt));
  }

  @Test
  public void getJustificationList_JustsNotFound() {
    Map<String, Object> claims = Map.of("id", "jwt-id");
    String token = Jwts.builder().setClaims(claims).compact();
    DecodedJWT jwt = JWT.decode(token);

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertNull(processor.getJustificationList(jwt));
  }

  @Test
  public void getJustificationList_JustsNotAList() {
    Map<String, Object> claims = Map.of("justs", "not a list");
    String token = Jwts.builder().setClaims(claims).compact();
    DecodedJWT jwt = JWT.decode(token);

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertNull(processor.getJustificationList(jwt));
  }

  @Test
  public void getJustificationList_JustsIsEmpty() {
    Map<String, Object> claims = Map.of("justs", List.of());
    String token = Jwts.builder().setClaims(claims).compact();
    DecodedJWT jwt = JWT.decode(token);

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertNull(processor.getJustificationList(jwt));
  }

  @Test
  public void getJustificationList_JustNotAMap() {
    Map<String, Object> claims = Map.of("justs", List.of("not a map"));
    String token = Jwts.builder().setClaims(claims).compact();
    DecodedJWT jwt = JWT.decode(token);

    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    assertNull(processor.getJustificationList(jwt));
  }

  private void verifyNoOp(AuditLogRequest request) throws Exception {
    JustificationProcessor processor = new JustificationProcessor(jvsClient);
    AuditLogRequest gotAuditLogReq = processor.process(request);

    assertSame(request, gotAuditLogReq);
    verifyNoInteractions(jvsClient);
  }
}
