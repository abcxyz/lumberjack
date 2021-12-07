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

package com.abcxyz.lumberjack.auditlogclient.processor;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.LoggingOptions;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/** Provides cloud logging-specific configuration. */
@Configuration
public class CloudLoggingProcessorConfiguration {

  @Bean
  Logging logging() {
    return LoggingOptions.getDefaultInstance().getService();
  }

  @Bean
  ObjectMapper mapper() {
    return new ObjectMapper();
  }
}
