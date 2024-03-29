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
import io.grpc.Metadata;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class SecurityContext {
  List<JwtSpecification> fromRawJwt;

  List<JwtSpecification> getFromRawJwt() {
    return fromRawJwt == null || fromRawJwt.isEmpty() ? Collections.emptyList() : fromRawJwt;
  }

  /** This is intended to be extended as we add more ways to specify security contexts. */
  public List<SecuritySpecification> getSecuritySpecifications() {
    List<SecuritySpecification> securitySpecifications = new ArrayList<>();
    securitySpecifications.addAll(getFromRawJwt());
    return securitySpecifications;
  }

  /** Use all configured security specifications in order to try to determine the principal. */
  public Optional<String> getPrincipal(Metadata headers) throws AuthorizationException {
    Optional<String> principal = Optional.empty();
    for (SecuritySpecification securitySpecification : getSecuritySpecifications()) {
      if (securitySpecification.isApplicable(headers)) {
        principal = securitySpecification.getPrincipal(headers);
        // Shortcut if we find a relevant principal.
        if (principal.isPresent()) {
          // TODO: do we want to handle multiple specs matching a single request?
          return principal;
        }
      }
    }
    return principal;
  }
}
