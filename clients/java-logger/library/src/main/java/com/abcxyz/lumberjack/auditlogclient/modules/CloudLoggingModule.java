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

// import java.util.Collections;

import org.threeten.bp.Duration;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.google.api.client.util.Strings;
import com.google.api.gax.batching.BatchingSettings;
import com.google.api.gax.retrying.RetrySettings;
// import com.google.cloud.logging.LogEntry;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.LoggingOptions;
// import com.google.cloud.logging.Severity;
// import com.google.cloud.logging.Payload.StringPayload;
import com.google.cloud.logging.Synchronicity;
import com.google.inject.AbstractModule;
import com.google.inject.Provides;
import com.google.inject.Singleton;

import lombok.extern.slf4j.Slf4j;

@Slf4j
public class CloudLoggingModule extends AbstractModule {
  @Provides
  @Singleton
  Logging logging(AuditLoggingConfiguration configuration) {
    LoggingOptions.Builder loggingOptionsBuilder = LoggingOptions.newBuilder();
    if (configuration.getBackend().cloudLoggingEnabled()
        && !Strings.isNullOrEmpty(configuration.getBackend().getCloudlogging().getProject())) {
      if (configuration.getBackend().getCloudlogging().useDefaultProject()) {
        throw new IllegalStateException("Cannot set cloud logging project if default is enabled.");
      }
      loggingOptionsBuilder.setProjectId(configuration.getBackend().getCloudlogging().getProject())
        .setBatchingSettings(BatchingSettings.newBuilder()
          .setElementCountThreshold(10L)
          .setDelayThreshold(Duration.ofSeconds(1L))
          .setRequestByteThreshold(2000L)
          .build())
        .setRetrySettings(RetrySettings.newBuilder()
          .setMaxRetryDelay(Duration.ofMillis(30000L))
          .setTotalTimeout(Duration.ofMillis(120000L))
          .setInitialRetryDelay(Duration.ofMillis(250L))
          .setRetryDelayMultiplier(1.0)
          .setInitialRpcTimeout(Duration.ofMillis(120000L))
          .setRpcTimeoutMultiplier(1.0)
          .setMaxRpcTimeout(Duration.ofMillis(120000L))
          .build());
    }
    Logging logging = loggingOptionsBuilder.build().getService();
    if (ConfigUtils.shouldFailClose(configuration.getLogMode())) {
      logging.setWriteSynchronicity(Synchronicity.SYNC);
    } else {
      logging.setWriteSynchronicity(Synchronicity.ASYNC);
      // warmup
      // try {
      //   LogEntry entry =
      //         LogEntry.newBuilder(StringPayload.of("warmup cloud logging"))
      //             .setSeverity(Severity.INFO)
      //             .setLogName(getClass().getSimpleName())
      //             .build();
      //   logging.write(Collections.singleton(entry));
      // } catch (Exception e) {
      //   log.warn("failed to warmup cloud logging", e);
      // }
    }
    return logging;
  }
}
