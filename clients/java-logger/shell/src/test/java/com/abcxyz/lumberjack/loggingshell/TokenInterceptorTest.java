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

package com.abcxyz.lumberjack.loggingshell;

import static io.jsonwebtoken.security.Keys.secretKeyFor;
import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.auditlogclient.LoggingClient;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import jakarta.servlet.http.HttpServletResponse;
import java.util.Map;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mock;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.web.method.HandlerMethod;

public class TokenInterceptorTest {
  private static final String AUTHORIZATION_HEADER_NAME = "Authorization";
  private final TokenInterceptor tokenInterceptor = new TokenInterceptor();
  private static final String TEST_EMAIL = "testEmail";
  private static final String REDACTED_SIGNATURE = "REDACTED_SIGNATURE";

  private static final String SIGNED_JWT_WITH_EMAIL_FIELD =
      Jwts.builder()
          .addClaims(Map.of(TokenInterceptor.JWT_EMAIL_FIELD_KEY, TEST_EMAIL))
          .signWith(secretKeyFor(SignatureAlgorithm.HS512), SignatureAlgorithm.HS512)
          .compact();

  private static final String REDACTED_SIGNATURE_JWT_WITH_EMAIL_FIELD =
      Jwts.builder().addClaims(Map.of(TokenInterceptor.JWT_EMAIL_FIELD_KEY, TEST_EMAIL)).compact()
          + REDACTED_SIGNATURE;

  private static final String UNSIGNED_JWT_WITH_EMAIL_FIELD =
      Jwts.builder().addClaims(Map.of(TokenInterceptor.JWT_EMAIL_FIELD_KEY, TEST_EMAIL)).compact();

  @MockitoBean private LoggingClient loggingClient;

  @Mock private HttpServletResponse mockResponse;
  @Mock private HandlerMethod mockHandler;
  private MockHttpServletRequest request;

  @BeforeEach
  void setUp() {
    request = new MockHttpServletRequest();
  }

  @Test
  void shouldAddEmailInfoFromSignedJwtToRequestAttributesWhenBearerIsLowercase() {
    request.addHeader(AUTHORIZATION_HEADER_NAME, "bearer " + SIGNED_JWT_WITH_EMAIL_FIELD);
    tokenInterceptor.preHandle(request, mockResponse, mockHandler);
    assertThat(request.getAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY))
        .isEqualTo(TEST_EMAIL);
  }

  @Test
  void shouldAddEmailInfoFromSignedJwtToRequestAttributes() {
    request.addHeader(AUTHORIZATION_HEADER_NAME, "Bearer " + SIGNED_JWT_WITH_EMAIL_FIELD);
    tokenInterceptor.preHandle(request, mockResponse, mockHandler);
    assertThat(request.getAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY))
        .isEqualTo(TEST_EMAIL);
  }

  @Test
  void shouldAddEmailInfoFromRedactedJwtToRequestAttributes() {
    request.addHeader(
        AUTHORIZATION_HEADER_NAME, "Bearer " + REDACTED_SIGNATURE_JWT_WITH_EMAIL_FIELD);
    tokenInterceptor.preHandle(request, mockResponse, mockHandler);
    assertThat(request.getAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY))
        .isEqualTo(TEST_EMAIL);
  }

  @Test
  void shouldAddEmailInfoFromUnsignedJwtToRequestAttributes() {
    request.addHeader(AUTHORIZATION_HEADER_NAME, "Bearer " + UNSIGNED_JWT_WITH_EMAIL_FIELD);
    tokenInterceptor.preHandle(request, mockResponse, mockHandler);
    assertThat(request.getAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY))
        .isEqualTo(TEST_EMAIL);
  }

  @Test
  void shouldThrowExceptionWhenAuthorizationHeaderIsMissing() {
    Exception ex =
        assertThrows(
            IllegalArgumentException.class,
            () -> tokenInterceptor.preHandle(request, mockResponse, mockHandler));
    assertThat(ex.getMessage()).isEqualTo("Missing Authorization header.");
  }

  @Test
  void shouldThrowExceptionWhenAuthorizationHeaderIsMalformed() {
    request.addHeader(AUTHORIZATION_HEADER_NAME, "invalid_token");
    Exception ex =
        assertThrows(
            IllegalArgumentException.class,
            () -> tokenInterceptor.preHandle(request, mockResponse, mockHandler));
    assertThat(ex.getMessage()).isEqualTo("Invalid Authorization header.");
  }

  @Test
  void shouldThrowExceptionWhenAuthorizationTokenIsMalformed() {
    request.addHeader(AUTHORIZATION_HEADER_NAME, "Bearer invalid_jwt");
    Exception ex =
        assertThrows(
            IllegalArgumentException.class,
            () -> tokenInterceptor.preHandle(request, mockResponse, mockHandler));
    assertThat(ex.getMessage()).isEqualTo("Invalid Authorization header.");
  }
}
