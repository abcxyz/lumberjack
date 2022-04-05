package com.abcxyz.lumberjack.auditlogclient.config;

import static org.assertj.core.api.Assertions.assertThat;

import com.abcxyz.lumberjack.auditlogclient.modules.AuditLoggingConfigurationModule;
import com.abcxyz.lumberjack.auditlogclient.modules.AuditLoggingModule;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import org.junit.jupiter.api.Test;

public class AuditLoggingConfigurationTest {
  @Test
  public void testMinimalConfiguration() throws IOException {
    AuditLoggingConfigurationModule configModule = new AuditLoggingConfigurationModule();
    AuditLoggingConfiguration config = configModule.auditLoggingConfiguration("minimal.yml");
    AuditLoggingModule module = new AuditLoggingModule();

    BackendContext expectedBackendContext = new BackendContext();
    LocalConfiguration local = new LocalConfiguration();
    local.setLogOutEnabled(true);
    expectedBackendContext.setLocal(local);
    assertThat(config.getBackend()).isEqualTo(expectedBackendContext);

    assertThat(config.getConditions()).isNull();
    assertThat(config.getRules().size()).isEqualTo(1);
    assertThat(config.getLogMode()).isEqualTo(LogMode.LOG_MODE_UNSPECIFIED);

    assertThat(module.backendContext(config)).isEqualTo(expectedBackendContext);
    assertThat(module.filters(config)).isEqualTo(new Filters());
  }

  @Test
  public void testMinimalConfiguration_Labels() throws IOException {
    AuditLoggingConfigurationModule configModule = new AuditLoggingConfigurationModule();
    AuditLoggingModule module = new AuditLoggingModule();
    AuditLoggingConfiguration config =
        configModule.auditLoggingConfiguration("minimal_with_labels.yml");

    BackendContext expectedBackendContext = new BackendContext();
    LocalConfiguration local = new LocalConfiguration();
    local.setLogOutEnabled(true);
    expectedBackendContext.setLocal(local);
    assertThat(config.getBackend()).isEqualTo(expectedBackendContext);

    assertThat(config.getConditions()).isNull();
    assertThat(config.getRules().size()).isEqualTo(1);
    assertThat(config.getLogMode()).isEqualTo(LogMode.LOG_MODE_UNSPECIFIED);

    assertThat(module.backendContext(config)).isEqualTo(expectedBackendContext);
    assertThat(module.filters(config)).isEqualTo(new Filters());

    Map<String, String> expectedLabels = new HashMap<>();
    expectedLabels.put("mylabel1", "myvalue1");
    expectedLabels.put("mylabel2", "myvalue2");
    assertThat(config.getLabels()).isEqualTo(expectedLabels);
  }
}
