package com.abcxyz.lumberjack.auditlogclient.modules;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.google.inject.AbstractModule;
import com.google.inject.Provides;
import com.google.inject.name.Named;
import java.io.IOException;
import java.io.InputStream;

public class AuditLoggingConfigurationModule extends AbstractModule {
  private static final String DEFAULT_CONFIG_LOCATION = "audit_logging.yml";
  private static final String CONFIG_ENV_KEY = "AUDIT_CLIENT_CONFIG_PATH";

  @Provides
  public AuditLoggingConfiguration auditLoggingConfiguration(
      @Named("AuditClientConfigName") String configName) {
    try {
      try (InputStream input = getClass().getClassLoader().getResourceAsStream(configName)) {
        ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
        return mapper.readValue(input, AuditLoggingConfiguration.class);
      }
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @Provides
  @Named("AuditClientConfigName")
  public String configName() {
    return System.getenv().containsKey(CONFIG_ENV_KEY)
        ? System.getenv().get(CONFIG_ENV_KEY)
        : DEFAULT_CONFIG_LOCATION;
  }
}
