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

package com.abcxyz.lumberjack.auditlogclient.modules;

import com.abcxyz.lumberjack.auditlogclient.LoggingClient;
import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.config.BackendContext;
import com.abcxyz.lumberjack.auditlogclient.config.Filters;
import com.abcxyz.lumberjack.auditlogclient.processor.RemoteProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.RuntimeInfoProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.ValidationProcessor;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.LoggingOptions;
import com.google.inject.AbstractModule;
import com.google.inject.Inject;
import com.google.inject.Provides;
import java.io.IOException;
import java.io.InputStream;
import java.util.List;

/**
 * This is the larger module intended to be consumed. It pulls in all related modules, and should
 * have all required dependencies to build audit log clients.
 */
public class AuditLoggingModule extends AbstractModule {
  private static final String DEFAULT_CONFIG_LOCATION = "application.yml";
  private static final String CONFIG_ENV_KEY = "AUDIT_LOGGING_CONFIGURATION";

  @Provides
  public AuditLoggingConfiguration auditLoggingConfiguration() {

    try {
      String fileLocation =
          System.getenv().containsKey(CONFIG_ENV_KEY)
              ? System.getenv().get(CONFIG_ENV_KEY)
              : DEFAULT_CONFIG_LOCATION;

      try (InputStream input = getClass().getClassLoader().getResourceAsStream(fileLocation)) {
        ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
        return mapper.readValue(input, AuditLoggingConfiguration.class);
      }
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @Provides
  public BackendContext backendContext(AuditLoggingConfiguration auditLoggingConfiguration) {
    return auditLoggingConfiguration.getBackend();
  }

  @Provides
  public Filters filters(AuditLoggingConfiguration auditLoggingConfiguration) {
    return auditLoggingConfiguration.getFilters();
  }

  @Provides
  Logging logging() {
    return LoggingOptions.getDefaultInstance().getService();
  }

  @Provides
  ObjectMapper mapper() {
    return new ObjectMapper();
  }

  @Provides
  @Inject
  public LoggingClient loggingClient(
      RuntimeInfoProcessor runtimeInfoProcessor,
      ValidationProcessor validationProcessor,
      RemoteProcessor remoteProcessor) {
    return new LoggingClient(
        List.of(validationProcessor), List.of(runtimeInfoProcessor), List.of(remoteProcessor));
  }

  @Override
  protected void configure() {
    install(new RuntimeInfoProcessorModule());
    install(new RemoteProcessorModule());
    install(new FilteringProcessorModule());
  }
}
