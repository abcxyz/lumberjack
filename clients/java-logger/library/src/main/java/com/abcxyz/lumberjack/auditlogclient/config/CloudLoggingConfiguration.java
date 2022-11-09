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

package com.abcxyz.lumberjack.auditlogclient.config;

import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.EqualsAndHashCode;
import lombok.Setter;

@Setter
@EqualsAndHashCode
public class CloudLoggingConfiguration {
  private static final String CLOUD_LOG_DEFAULT_KEY =
      "AUDIT_CLIENT_BACKEND_CLOUDLOGGING_DEFAULT_PROJECT";
  private static final String CLOUD_LOG_PROJECT_KEY = "AUDIT_CLIENT_BACKEND_CLOUDLOGGING_PROJECT";

  @JsonProperty("default_project")
  private boolean defaultProject;

  private String project;

  public String getProject() {
    return ConfigUtils.getEnvOrDefault(CLOUD_LOG_PROJECT_KEY, project);
  }

  public boolean useDefaultProject() {
    return ConfigUtils.getEnvOrDefault(CLOUD_LOG_DEFAULT_KEY, defaultProject);
  }
}
