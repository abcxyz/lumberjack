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

package com.abcxyz.lumberjack.auditlogclient.utils;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;

public class ConfigUtils {
  private ConfigUtils() {
    // no-op
  }

  /**
   * Returns whether we should fail close on errors. Unspecified (LOG_MODE_UNSPECIFIED) is handled
   * equivalently to BEST_EFFORT, which is to not fail close.
   */
  public static boolean shouldFailClose(LogMode logMode) {
    return logMode.equals(LogMode.FAIL_CLOSE);
  }

  /** Returns the env var or the default value if unspecified. */
  public static String getEnvOrDefault(String envKey, String defaultValue) {
    return System.getenv().getOrDefault(envKey, defaultValue);
  }

  /** Returns the env var boolean value or the default value if unspecified. */
  public static boolean getEnvOrDefault(String envKey, boolean defaultValue) {
    return System.getenv().containsKey(envKey)
        ? Boolean.valueOf(System.getenv().get(envKey))
        : defaultValue;
  }
}
