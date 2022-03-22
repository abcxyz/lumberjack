package com.abcxyz.lumberjack.auditlogclient.config;

import com.fasterxml.jackson.annotation.JsonProperty;

public class LocalConfiguration {
  private static final String LOG_OUT_ENABLED_KEY = "AUDIT_CLIENT_LOG_OUT_ENABLED";

  @JsonProperty("log_out_enabled")
  private boolean logOutEnabled;

  public boolean logOutEnabled() {
    if (System.getenv().containsKey(LOG_OUT_ENABLED_KEY)) {
      return Boolean.valueOf(System.getenv().get(LOG_OUT_ENABLED_KEY));
    }
    return logOutEnabled;
  }
}
