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

package com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.name.Named;

/**
 * CloudFunctionManager provides functionality for getting run time info for processes running on
 * cloud function.
 */
public class CloudFunctionManager {

  private final String functionTarget, functionSigType, service, revision;
  private final RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Inject
  public CloudFunctionManager(
      @Named("FUNCTION_SIGNATURE_TYPE") final String functionSigType,
      @Named("FUNCTION_TARGET") final String functionTarget,
      @Named("K_SERVICE") final String service,
      @Named("K_REVISION") final String revision,
      RuntimeInfoCommonUtils runtimeInfoCommonUtils) {
    this.functionTarget = functionTarget;
    this.functionSigType = functionSigType;
    this.service = service;
    this.revision = revision;
    this.runtimeInfoCommonUtils = runtimeInfoCommonUtils;
  }

  /**
   * Detect if current process is being run on an cloud functions. Based on
   * https://cloud.google.com/functions/docs/configuring/env-var#newer_runtimes
   */
  public MonitoredResource detectCloudFunction() {
    return MonitoredResource.newBuilder()
        .setType("cloud_function")
        .putLabels("project_id", runtimeInfoCommonUtils.getProjectId())
        .putLabels("function_name", service)
        .putLabels("region", runtimeInfoCommonUtils.getRegion())
        .build();
  }

  public boolean isCloudFunction() {
    return !runtimeInfoCommonUtils.isNullOrBlank(functionTarget)
        && !runtimeInfoCommonUtils.isNullOrBlank(functionSigType)
        && !runtimeInfoCommonUtils.isNullOrBlank(service)
        && !runtimeInfoCommonUtils.isNullOrBlank(revision);
  }
}
