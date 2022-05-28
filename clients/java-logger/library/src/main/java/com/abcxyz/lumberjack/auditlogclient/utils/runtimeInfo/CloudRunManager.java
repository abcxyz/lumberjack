package com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.name.Named;

/**
 * CloudRunManager provides functionality for getting run time info for processes running on cloud
 * Run.
 */
public class CloudRunManager {

  private final String configuration;
  private final String service;
  private final String revision;
  private final RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Inject
  public CloudRunManager(
      @Named("K_CONFIGURATION") final String configuration,
      @Named("K_SERVICE") final String service,
      @Named("K_REVISION") final String revision,
      RuntimeInfoCommonUtils runtimeInfoCommonUtils) {
    this.configuration = configuration;
    this.service = service;
    this.revision = revision;
    this.runtimeInfoCommonUtils = runtimeInfoCommonUtils;
  }

  public MonitoredResource detectCloudRunResource() {
    return MonitoredResource.newBuilder()
        .setType("cloud_run_revision")
        .putLabels("project_id", runtimeInfoCommonUtils.getProjectId())
        .putLabels("location", runtimeInfoCommonUtils.getRegion())
        .putLabels("service_name", service)
        .putLabels("revision_name", revision)
        .putLabels("configuration_name", configuration)
        .build();
  }

  /**
   * Detect if current process is being run on an cloud functions. Based on
   * https://cloud.google.com/anthos/run/docs/reference/container-contract#env-vars
   */
  public boolean isCloudRun() {
    return !runtimeInfoCommonUtils.isNullOrBlank(configuration)
        && !runtimeInfoCommonUtils.isNullOrBlank(service)
        && !runtimeInfoCommonUtils.isNullOrBlank(revision);
  }
}
