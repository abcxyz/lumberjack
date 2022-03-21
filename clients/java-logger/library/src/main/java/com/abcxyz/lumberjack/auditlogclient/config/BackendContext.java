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
import com.google.api.client.util.Strings;
import lombok.Data;

/**
 * Contains configuration pertaining to RemoteProcessors. Each value defaults to the value in YAML
 * configuration, but may be overridden using environment variables.
 */
@Data
public class BackendContext {
  private static final String ADDRESS_ENV_KEY = "AUDIT_CLIENT_BACKEND_ADDRESS";
  private static final String AUTH_AUDIENCE_ENV_KEY = "AUDIT_CLIENT_BACKEND_AUTH_AUDIENCE";
  private static final String IMPERSONATE_ACC_ENV_KEY = "AUDIT_CLIENT_BACKEND_IMPERSONATE_ACCOUNT";
  private static final String INSECURE_ENABLED_ENV_KEY = "AUDIT_CLIENT_BACKEND_INSECURE_ENABLED";
  private static final String LOCAL_LOGGIN_ENABLED_ENV_KEY = "AUDIT_CLIENT_LOCAL_LOGGING_ENABLED";

  private String address;

  @JsonProperty("auth_audience")
  private String authAudience;

  @JsonProperty("impersonate_account")
  private String impersonateAccount;

  @JsonProperty("insecure_enabled")
  private boolean insecureEnabled; // meant for use in unit tests only

  @JsonProperty("local_logging_enabled")
  private boolean localLoggingEnabled;

  /**
   * If address has been set, then we should log to the remote processor at that address.
   */
  public boolean remoteEnabled() {
    return !Strings.isNullOrEmpty(getAddress());
  }

  public boolean localLoggingEnabled() {
    if (System.getenv().containsKey(LOCAL_LOGGIN_ENABLED_ENV_KEY)) {
      return Boolean.valueOf(System.getenv().get(LOCAL_LOGGIN_ENABLED_ENV_KEY));
    }
    return localLoggingEnabled;
  }

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
