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

package com.abcxyz.lumberjack.auditlogclient.modules;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.google.inject.AbstractModule;
import com.google.inject.Provides;
import com.google.inject.name.Named;
import java.io.IOException;
import java.io.InputStream;

public class AuditLoggingConfigurationModule extends AbstractModule {
  private static final String DEFAULT_CONFIG_LOCATION = "audit_logging.yml";
  private static final String CONFIG_ENV_KEY = "AUDIT_CLIENT_CONFIG_NAME";

  @Provides
  public AuditLoggingConfiguration auditLoggingConfiguration(
      @Named("AuditClientConfigName") String configName) {
    try {
      try (InputStream input = getClass().getClassLoader().getResourceAsStream(configName)) {
        ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
        mapper.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
        return mapper.readValue(input, AuditLoggingConfiguration.class);
      }
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @Provides
  @Named("AuditClientConfigName")
  public String configName() {
    return ConfigUtils.getEnvOrDefault(CONFIG_ENV_KEY, DEFAULT_CONFIG_LOCATION);
  }
}
