package com.abcxyz.lumberjack.auditlogclient.config;

import static org.assertj.core.api.Assertions.assertThat;

import com.abcxyz.lumberjack.auditlogclient.modules.AuditLoggingModule;
import java.io.IOException;
import org.junit.jupiter.api.Test;

public class AuditLoggingConfigurationTest {
  @Test
  public void testMinimalConfiguration() throws IOException {
    AuditLoggingModule module = new AuditLoggingModule();
    AuditLoggingConfiguration config = module.auditLoggingConfiguration("minimal.yml");
    assertThat(config.getBackend()).isNull();
    assertThat(config.getConditions()).isNull();
    assertThat(config.getRules().size()).isEqualTo(1);

    assertThat(module.backendContext(config)).isEqualTo(new BackendContext());
    assertThat(module.filters(config)).isEqualTo(new Filters());
  }
}
