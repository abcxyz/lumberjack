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
