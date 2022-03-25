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
import com.abcxyz.lumberjack.auditlogclient.LoggingClientBuilder;
import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.config.BackendContext;
import com.abcxyz.lumberjack.auditlogclient.config.Filters;
import com.abcxyz.lumberjack.auditlogclient.processor.RemoteProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.RuntimeInfoProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.ValidationProcessor;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.google.api.client.util.Strings;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.LoggingOptions;
import com.google.inject.AbstractModule;
import com.google.inject.Inject;
import com.google.inject.Provides;
import com.google.inject.name.Named;
import java.io.IOException;
import java.io.InputStream;
import java.util.List;

/**
 * This is the larger module intended to be consumed. It pulls in all related modules, and should
 * have all required dependencies to build audit log clients.
 */
public class AuditLoggingModule extends AbstractModule {
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

  @Provides
  public BackendContext backendContext(AuditLoggingConfiguration auditLoggingConfiguration) {
      return auditLoggingConfiguration.getBackend();
  }

  @Provides
  public Filters filters(AuditLoggingConfiguration auditLoggingConfiguration) {
    return auditLoggingConfiguration.getFilters();
  }

  @Provides
  Logging logging(AuditLoggingConfiguration configuration) {
    LoggingOptions loggingOptions = LoggingOptions.getDefaultInstance();
    if (configuration.getBackend().cloudLoggingEnabled()
        && !Strings.isNullOrEmpty(configuration.getBackend().getCloudlogging().getProject())) {
      if (configuration.getBackend().getCloudlogging().useDefaultProject()){
        throw new IllegalStateException("Cannot set cloud logging project if default is enabled.");
      }
      loggingOptions = loggingOptions.toBuilder()
          .setProjectId(configuration.getBackend().getCloudlogging().getProject())
          .build();
    }
    return loggingOptions.getService();
  }

  @Provides
  ObjectMapper mapper() {
    return new ObjectMapper();
  }

  @Provides
  @Inject
  public LoggingClient loggingClient(LoggingClientBuilder builder) {
    return builder.withDefaultProcessors().build();
  }

  @Override
  protected void configure() {
    install(new RuntimeInfoProcessorModule());
    install(new RemoteProcessorModule());
    install(new FilteringProcessorModule());
  }
}
