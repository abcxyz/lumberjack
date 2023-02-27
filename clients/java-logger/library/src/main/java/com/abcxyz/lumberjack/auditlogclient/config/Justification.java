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

package com.abcxyz.lumberjack.auditlogclient.config;

import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class Justification {
  private static final String JUSTIFICATION_PUBLIC_KEYS_ENDPOINT_ENV_KEY =
      "AUDIT_CLIENT_JUSTIFICATION_PUBLIC_KEYS_ENDPOINT";
  private static final String JUSTIFICATION_ENABLED_ENV_KEY = "AUDIT_CLIENT_JUSTIFICATION_ENABLED";
  private static final String JUSTIFICATION_ALLOW_BREAKGLASS_ENV_KEY =
      "AUDIT_CLIENT_JUSTIFICATION_ALLOW_BREAKGLASS";

  @JsonProperty("public_keys_endpoint")
  private String publicKeysEndpoint;

  private boolean enabled;

  // Default value is false, this field will be ignored if justification is not enabled.
  @JsonProperty("allow_breakglass")
  private boolean allowBreakglass;

  public String getPublicKeysEndpoint() {
    return ConfigUtils.getEnvOrDefault(
        JUSTIFICATION_PUBLIC_KEYS_ENDPOINT_ENV_KEY, publicKeysEndpoint);
  }

  public boolean isEnabled() {
    return ConfigUtils.getEnvOrDefault(JUSTIFICATION_ENABLED_ENV_KEY, enabled);
  }

  public boolean allowBreakglass() {
    return ConfigUtils.getEnvOrDefault(JUSTIFICATION_ALLOW_BREAKGLASS_ENV_KEY, allowBreakglass);
  }

  public void validate() {
    if (isEnabled() && (publicKeysEndpoint == null || publicKeysEndpoint.isEmpty())) {
      throw new IllegalArgumentException(
          "public_keys_endpoint must be specified when justification is enabled");
    }
  }
}
