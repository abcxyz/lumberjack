package com.abcxyz.lumberjack.auditlogclient.utils;

import static org.assertj.core.api.Assertions.assertThat;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;
import org.junit.jupiter.api.Test;

public class ConfigUtilsTest {
  @Test
  public void testShouldFailClose() {
    assertThat(ConfigUtils.shouldFailClose(LogMode.BEST_EFFORT)).isFalse();
    assertThat(ConfigUtils.shouldFailClose(LogMode.FAIL_CLOSE)).isTrue();
    assertThat(ConfigUtils.shouldFailClose(LogMode.LOG_MODE_UNSPECIFIED)).isFalse();
  }
}
