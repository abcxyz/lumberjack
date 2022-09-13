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

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class Justification {
  private static final String JUSTIFICATION_PUBLIC_KEYS_ENDPOINT_ENV_KEY =
      "AUDIT_CLIENT_JUSTIFICATION_PUBLIC_KEYS_ENDPOINT";
  private static final String JUSTIFICATION_ENABLED_ENV_KEY = "AUDIT_CLIENT_JUSTIFICATION_ENABLED";

  @JsonProperty("public_keys_endpoint")
  private String publicKeysEndpoint = "localhost:8080";

  @JsonProperty("enabled")
  private boolean enabled;

  public String getPublicKeysEndpoint() {
    return System.getenv()
        .getOrDefault(JUSTIFICATION_PUBLIC_KEYS_ENDPOINT_ENV_KEY, publicKeysEndpoint);
  }

  public boolean isEnabled() {
    if (System.getenv().containsKey(JUSTIFICATION_ENABLED_ENV_KEY)) {
      return Boolean.valueOf(System.getenv().get(JUSTIFICATION_ENABLED_ENV_KEY));
    }
    return enabled;
  }
}
