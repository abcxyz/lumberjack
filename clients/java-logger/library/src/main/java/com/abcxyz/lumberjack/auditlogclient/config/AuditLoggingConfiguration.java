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

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogMode;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonSetter;
import com.fasterxml.jackson.annotation.Nulls;
import java.util.List;
import java.util.Map;
import lombok.Data;
import lombok.NoArgsConstructor;

/**
 * This represents the entire YAML file, and will be the target for deserialization from the file.
 */
@Data
@NoArgsConstructor
public class AuditLoggingConfiguration {
  private String version;
  private BackendContext backend;

  @JsonProperty("log_mode")
  private LogMode logMode;

  @JsonProperty("condition")
  private ConditionConfig conditions;

  private List<Selector> rules;

  @JsonSetter(nulls = Nulls.FAIL)
  @JsonProperty("security_context")
  private SecurityContext securityContext;

  @JsonProperty("justification")
  private Justification justification = new Justification();

  @JsonProperty(value = "breakglass_allowed")
  private boolean breakglassAllowed = false;

  private Map<String, String> labels;

  public Filters getFilters() {
    return conditions == null ? new Filters() : conditions.getFilters();
  }

  public LogMode getLogMode() {
    return logMode == null ? LogMode.LOG_MODE_UNSPECIFIED : logMode;
  }

  public Justification getJustificaiton() {
    if (justification != null) {
      justification.validate();
    }
    return justification;
  }

  @Data
  private class ConditionConfig {
    @JsonProperty("regex")
    private Filters filters;
  }

  public BackendContext getBackend() {
    if (backend == null) {
      // if no backend context is specified, default to local logging.
      backend = new BackendContext();
      LocalConfiguration local = new LocalConfiguration();
      local.setLogOutEnabled(true);
      backend.setLocal(local);
    }

    return backend;
  }
}
