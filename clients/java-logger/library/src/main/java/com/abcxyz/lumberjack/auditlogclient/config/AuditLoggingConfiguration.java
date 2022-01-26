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

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.annotation.JsonSetter;
import com.fasterxml.jackson.annotation.Nulls;
import java.util.List;
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

  @JsonProperty("condition")
  private ConditionConfig conditions;

  private List<Selector> rules;

  @JsonSetter(nulls = Nulls.FAIL)
  @JsonProperty("security_context")
  private SecurityContext securityContext;

  public Filters getFilters() {
    return conditions == null ?
        new Filters() :
        conditions.getFilters();
  }

  @Data
  private class ConditionConfig {
    @JsonProperty("regex")
    private Filters filters;
  }
}
