package com.abcxyz.lumberjack.auditlogclient.modules;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.google.api.client.util.Strings;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.LoggingOptions;
import com.google.cloud.logging.Synchronicity;
import com.google.inject.AbstractModule;
import com.google.inject.Provides;

public class CloudLoggingModule extends AbstractModule {
  @Provides
  Logging logging(AuditLoggingConfiguration configuration) {
    LoggingOptions.Builder loggingOptionsBuilder = LoggingOptions.newBuilder();
    if (configuration.getBackend().cloudLoggingEnabled()
        && !Strings.isNullOrEmpty(configuration.getBackend().getCloudlogging().getProject())) {
      if (configuration.getBackend().getCloudlogging().useDefaultProject()) {
        throw new IllegalStateException("Cannot set cloud logging project if default is enabled.");
      }
      loggingOptionsBuilder.setProjectId(configuration.getBackend().getCloudlogging().getProject());
    }
    Logging logging = loggingOptionsBuilder.build().getService();
    if (ConfigUtils.shouldFailClose(configuration.getLogMode())) {
      logging.setWriteSynchronicity(Synchronicity.SYNC);
    } else {
      logging.setWriteSynchronicity(Synchronicity.ASYNC);
    }
    return logging;
  }
}
