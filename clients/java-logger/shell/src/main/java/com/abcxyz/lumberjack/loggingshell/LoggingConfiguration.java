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

package com.abcxyz.lumberjack.loggingshell;

import com.abcxyz.lumberjack.auditlogclient.LoggingClient;
// import com.abcxyz.lumberjack.auditlogclient.LoggingClientBuilder;
import com.abcxyz.lumberjack.auditlogclient.modules.AuditLoggingModule;
import com.google.inject.Guice;
import com.google.inject.Injector;
import java.time.Clock;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/** Provides logging-specific configuration. */
@Configuration
public class LoggingConfiguration {
  @Bean
  LoggingClient loggingClient() {
    Injector injector = Guice.createInjector(new AuditLoggingModule());
    // LoggingClientBuilder builder = injector.getInstance(LoggingClientBuilder.class);
    // return builder.withDefaultProcessors().build();
    return injector.getInstance(LoggingClient.class);
  }

  @Bean
  Clock clock() {
    return Clock.systemUTC();
  }
}
