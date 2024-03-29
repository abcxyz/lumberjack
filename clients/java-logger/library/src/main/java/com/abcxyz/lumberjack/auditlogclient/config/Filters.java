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

package com.abcxyz.lumberjack.auditlogclient.config;

import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import lombok.Data;

@Data
public class Filters {
  static final String PRINCIPAL_INCLUDE_ENV_KEY = "AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE";
  static final String PRINCIPAL_EXCLUDE_ENV_KEY = "AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_EXCLUDE";

  private String principalInclude;

  private String principalExclude;

  public String getPrincipalInclude() {
    return ConfigUtils.getEnvOrDefault(PRINCIPAL_INCLUDE_ENV_KEY, principalInclude);
  }

  public String getPrincipalExclude() {
    return ConfigUtils.getEnvOrDefault(PRINCIPAL_EXCLUDE_ENV_KEY, principalExclude);
  }
}
