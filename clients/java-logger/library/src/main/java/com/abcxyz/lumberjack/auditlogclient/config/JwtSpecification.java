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

import com.abcxyz.lumberjack.auditlogclient.exceptions.AuthorizationException;
import com.auth0.jwt.JWT;
import com.auth0.jwt.exceptions.JWTDecodeException;
import com.auth0.jwt.interfaces.Claim;
import com.auth0.jwt.interfaces.DecodedJWT;
import io.grpc.Metadata;
import java.util.Map;
import java.util.Optional;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class JwtSpecification implements SecuritySpecification {
  private static final String EMAIL_KEY = "email";

  private String key;
  private String prefix;

  private JwksSpecification jwks;

  @Override
  public boolean isApplicable(Metadata headers) {
    Metadata.Key<String> metadataKey = Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
    return headers.containsKey(metadataKey);
  }

  @Override
  public Optional<String> getPrincipal(Metadata headers) throws AuthorizationException {
    if (!isApplicable(headers)) {
      return Optional.empty();
    }
    Metadata.Key<String> metadataKey = Metadata.Key.of(key, Metadata.ASCII_STRING_MARSHALLER);
    String idToken = headers.get(metadataKey);
    if (prefix != null && idToken.toLowerCase().startsWith(prefix.toLowerCase())) {
      idToken = idToken.substring(prefix.length());
    }
    try {
      DecodedJWT jwt = JWT.decode(idToken);
      Map<String, Claim> claims = jwt.getClaims();
      if (!claims.containsKey(EMAIL_KEY)) {
        return Optional.empty();
      }
      String principal = claims.get(EMAIL_KEY).asString();
      return Optional.of(principal);
    } catch (JWTDecodeException e) {
      // invalid token
      throw new AuthorizationException("JWT Token was invalid", e);
    }
  }
}
