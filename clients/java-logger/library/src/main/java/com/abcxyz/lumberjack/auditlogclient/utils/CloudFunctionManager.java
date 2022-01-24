package com.abcxyz.lumberjack.auditlogclient.utils;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.name.Named;

public class CloudFunctionManager {

  private final String functionTarget, functionSigType, service, revision;
  private final RuntimeInfoUtils runtimeInfoUtils;

  @Inject
  public CloudFunctionManager(@Named("FUNCTION_SIGNATURE_TYPE") final String functionSigType,
      @Named("FUNCTION_TARGET") final String functionTarget,
      @Named("K_SERVICE") final String service, @Named("K_REVISION") final String revision,
      RuntimeInfoUtils runtimeInfoUtils) {
    this.functionTarget = functionTarget;
    this.functionSigType = functionSigType;
    this.service = service;
    this.revision = revision;
    this.runtimeInfoUtils = runtimeInfoUtils;
  }

  /**
   * Detect if current process is being run on an cloud functions. Based on
   * https://cloud.google.com/functions/docs/configuring/env-var#newer_runtimes
   */
  public MonitoredResource detectCloudFunction() {
    return MonitoredResource.newBuilder()
        .setType("cloud_function")
        .putLabels("project_id", runtimeInfoUtils.getProjectId())
        .putLabels("function_name", service)
        .putLabels("region", runtimeInfoUtils.getRegion())
        .build();
  }

  public boolean isCloudFunction() {
    return
        (!runtimeInfoUtils.isNullOrBlank(functionTarget)
            && !runtimeInfoUtils.isNullOrBlank(functionSigType)
            && !runtimeInfoUtils.isNullOrBlank(service)
            && !runtimeInfoUtils.isNullOrBlank(revision));
  }

}
