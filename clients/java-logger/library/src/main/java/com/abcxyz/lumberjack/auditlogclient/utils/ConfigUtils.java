package com.abcxyz.lumberjack.auditlogclient.utils;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;

public class ConfigUtils {
  private ConfigUtils() {
    // no-op
  }

  /**
   * Returns whether we should fail close on errors. Unspecified (LOG_MODE_UNSPECIFIED)
   * is handled equivalently to BEST_EFFORT, which is to not fail close.
   */
  public static boolean shouldFailClose(LogMode logMode) {
    return logMode.equals(LogMode.FAIL_CLOSE);
  }
}
