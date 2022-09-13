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

import com.abcxyz.jvs.JVSClientBuilder;
import com.abcxyz.jvs.JvsClient;
import com.abcxyz.lumberjack.auditlogclient.LoggingClient;
import com.abcxyz.lumberjack.auditlogclient.LoggingClientBuilder;
import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.config.BackendContext;
import com.abcxyz.lumberjack.auditlogclient.config.Filters;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.inject.AbstractModule;
import com.google.inject.Inject;
import com.google.inject.Provides;
import java.time.Clock;

/**
 * This is the larger module intended to be consumed. It pulls in all related modules, and should
 * have all required dependencies to build audit log clients.
 */
public class AuditLoggingModule extends AbstractModule {
  @Provides
  public BackendContext backendContext(AuditLoggingConfiguration auditLoggingConfiguration) {
    return auditLoggingConfiguration.getBackend();
  }

  @Provides
  public Filters filters(AuditLoggingConfiguration auditLoggingConfiguration) {
    return auditLoggingConfiguration.getFilters();
  }

  @Provides
  public JvsClient jvsClient(AuditLoggingConfiguration auditLoggingConfiguration) {
    return new JVSClientBuilder()
        .withJvsEndpoint(auditLoggingConfiguration.getJustification().getPublicKeysEndpoint())
        .withAllowBreakglass(auditLoggingConfiguration.isBreakglassAllowed())
        .build();
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

  @Provides
  Clock clock() {
    return Clock.systemUTC();
  }

  @Override
  protected void configure() {
    install(new RuntimeInfoProcessorModule());
    install(new RemoteProcessorModule());
    install(new FilteringProcessorModule());
    install(new AuditLoggingConfigurationModule());
    install(new CloudLoggingModule());
  }
}
