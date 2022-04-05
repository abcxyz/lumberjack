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
  void whenLogModeIsUnspecifiedLoggingIsAsync() {
    AuditLoggingConfigurationTestModule.logMode = LogMode.LOG_MODE_UNSPECIFIED;
    Logging logging = injector().getInstance(Logging.class);
    assertThat(logging.getWriteSynchronicity()).isEqualTo(Synchronicity.ASYNC);
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
      config.setLogMode(logMode);
      return config;
    }
  }
}
