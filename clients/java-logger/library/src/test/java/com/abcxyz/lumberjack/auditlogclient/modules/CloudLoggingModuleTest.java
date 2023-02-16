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

import static org.assertj.core.api.Assertions.assertThat;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.Synchronicity;
import com.google.inject.AbstractModule;
import com.google.inject.Guice;
import com.google.inject.Injector;
import com.google.inject.Provides;
import org.junit.jupiter.api.Test;

class CloudLoggingModuleTest {
  private static Injector injector() {
    return Guice.createInjector(
        new CloudLoggingModule(), new AuditLoggingConfigurationTestModule());
  }

  @Test
  void whenLogModeIsBestEffortLoggingIsAsync() {
    AuditLoggingConfigurationTestModule.logMode = LogMode.BEST_EFFORT;
    Logging logging = injector().getInstance(Logging.class);
    assertThat(logging.getWriteSynchronicity()).isEqualTo(Synchronicity.ASYNC);
  }

  @Test
  void whenLogModeIsUnspecifiedLoggingIsSync() {
    AuditLoggingConfigurationTestModule.logMode = LogMode.LOG_MODE_UNSPECIFIED;
    Logging logging = injector().getInstance(Logging.class);
    assertThat(logging.getWriteSynchronicity()).isEqualTo(Synchronicity.SYNC);
  }

  @Test
  void whenLogModeIsFailCloseLoggingIsSync() {
    AuditLoggingConfigurationTestModule.logMode = LogMode.FAIL_CLOSE;
    Logging logging = injector().getInstance(Logging.class);
    assertThat(logging.getWriteSynchronicity()).isEqualTo(Synchronicity.SYNC);
  }

  private static class AuditLoggingConfigurationTestModule extends AbstractModule {
    private static LogMode logMode;

    @Provides
    public AuditLoggingConfiguration auditLoggingConfiguration() {
      AuditLoggingConfiguration config = new AuditLoggingConfiguration();
      config.getBackend().getCloudlogging().setProject("foo");
      config.setLogMode(logMode);
      return config;
    }
  }
}
