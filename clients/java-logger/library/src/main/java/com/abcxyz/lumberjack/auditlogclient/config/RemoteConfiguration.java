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

  private String authAudience;

  private String impersonateAccount;

  private boolean insecureEnabled; // meant for use in unit tests only

  public String getAddress() {
    return ConfigUtils.getEnvOrDefault(ADDRESS_ENV_KEY, address);
  }

  public String getAuthAudience() {
    return ConfigUtils.getEnvOrDefault(AUTH_AUDIENCE_ENV_KEY, authAudience);
  }

  public String getImpersonateAccount() {
    return ConfigUtils.getEnvOrDefault(IMPERSONATE_ACC_ENV_KEY, impersonateAccount);
  }

  public boolean getInsecureEnabled() {
    return ConfigUtils.getEnvOrDefault(INSECURE_ENABLED_ENV_KEY, insecureEnabled);
  }
}
