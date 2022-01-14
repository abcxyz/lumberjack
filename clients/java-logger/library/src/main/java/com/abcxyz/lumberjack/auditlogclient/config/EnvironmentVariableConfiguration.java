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

import com.google.inject.Inject;
import com.google.inject.name.Named;
import lombok.AllArgsConstructor;
import lombok.Data;

/**
 * This contains environment variables that are only specified through env variables, and cannot be
 * found in the YAML file.
 */
@Data
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class EnvironmentVariableConfiguration {
  @Named("K_CONFIGURATION")
  private final String kConfiguration;

  @Named("K_SERVICE")
  private final String kService;

  @Named("K_REVISION")
  private final String kRevision;

  @Named("GOOGLE_CLOUD_PROJECT")
  private final String googleCloudProjectId;

  @Named("GAE_SERVICE")
  private final String gaeService;

  @Named("GAE_VERSION")
  private final String gaeVersion;

  @Named("GAE_INSTANCE")
  private final String gaeInstance;

  @Named("GAE_RUNTIME")
  private final String gaeRuntime;

  @Named("FUNCTION_NAME")
  private final String functionName;

  @Named("FUNCTION_REGION")
  private final String functionRegion;

  @Named("ENTRY_POINT")
  private final String functionPoint;

  @Named("FUNCTION_TARGET")
  private final String functionTarget;

  @Named("FUNCTION_SIGNATURE_TYPE")
  private final String functionSignatureType;

  @Named("HOSTNAME")
  private final String hostname;

  @Named("CONTAINER_NAME")
  private final String containerName;
}
