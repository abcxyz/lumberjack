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

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class RemoteConfiguration {
  private static final String ADDRESS_ENV_KEY = "AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS";
  private static final String AUTH_AUDIENCE_ENV_KEY = "AUDIT_CLIENT_BACKEND_AUTH_AUDIENCE";
  private static final String IMPERSONATE_ACC_ENV_KEY =
      "AUDIT_CLIENT_BACKEND_REMOTE_IMPERSONATE_ACCOUNT";
  private static final String INSECURE_ENABLED_ENV_KEY =
      "AUDIT_CLIENT_BACKEND_REMOTE_INSECURE_ENABLED";

  private String address;

  @JsonProperty("auth_audience")
  private String authAudience;

  @JsonProperty("impersonate_account")
  private String impersonateAccount;

  @JsonProperty("insecure_enabled")
  private boolean insecureEnabled; // meant for use in unit tests only

  public String getAddress() {
    return System.getenv().getOrDefault(ADDRESS_ENV_KEY, address);
  }

  public String getAuthAudience() {
    return System.getenv().getOrDefault(AUTH_AUDIENCE_ENV_KEY, authAudience);
  }

  public String getImpersonateAccount() {
    return System.getenv().getOrDefault(IMPERSONATE_ACC_ENV_KEY, impersonateAccount);
  }

  public boolean getInsecureEnabled() {
    if (System.getenv().containsKey(INSECURE_ENABLED_ENV_KEY)) {
      return Boolean.valueOf(System.getenv().get(INSECURE_ENABLED_ENV_KEY));
    }
    return insecureEnabled;
  }
}
