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

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwt;
import io.jsonwebtoken.Jwts;
import javax.annotation.Nonnull;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

/** Intercepts the request and extracts the user email contained in the JWT used for auth. */
@Component
public class TokenInterceptor implements HandlerInterceptor {
  static final String AUTHORIZATION_HEADER_NAME = "Authorization";
  static final String JWT_EMAIL_FIELD_KEY = "email";
  static final String INTERCEPTOR_USER_EMAIL_KEY = "user_email";
  private static final String IAP_USER_EMAIL_HEADER = "X-Goog-Authenticated-User-Email";

  @Override
  public boolean preHandle(
      HttpServletRequest request, @Nonnull HttpServletResponse response, @Nonnull Object handler) {
    // Check if user email available via IAP.
    String iapUserEmail = request.getHeader(IAP_USER_EMAIL_HEADER);
    if (iapUserEmail != null && !iapUserEmail.isBlank()) {
      request.setAttribute(INTERCEPTOR_USER_EMAIL_KEY, iapUserEmail);
      return true;
    }

    // If user's email not available via IAP, parse the JWT to obtain the user email.
    String bearerAccessToken = parseBearerAccessToken(request.getHeader(AUTHORIZATION_HEADER_NAME));
    Jwt<?, Claims> jwt = Jwts.parserBuilder().build().parseClaimsJwt(bearerAccessToken);
    Claims claims = jwt.getBody();
    request.setAttribute(INTERCEPTOR_USER_EMAIL_KEY, claims.get(JWT_EMAIL_FIELD_KEY, String.class));
    return true;
  }

  private String parseBearerAccessToken(String authHeader) {
    if (authHeader == null || authHeader.isEmpty()) {
      throw new IllegalArgumentException("Missing Authorization header.");
    }
    String[] splitAuthHeader = authHeader.split("\\s+");
    if (splitAuthHeader.length != 2 || !splitAuthHeader[0].equalsIgnoreCase("bearer")) {
      throw new IllegalArgumentException("Invalid Authorization header.");
    }
    String token = splitAuthHeader[1];
    int signatureStartIdx = token.lastIndexOf('.');
    if (signatureStartIdx < 0) {
      throw new IllegalArgumentException("Invalid Authorization header.");
    }
    // Keep the last "." in order for JWT parsers to be able to parse.
    return token.substring(0, signatureStartIdx + 1);
  }
}
