package com.abcxyz.lumberjack.auditlogclient.utils;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.name.Named;

public class CloudRunManager {

  private final String configuration;
  private final String service;
  private final String revision;
  private final RuntimeInfoUtils runtimeInfoUtils;

  @Inject
  public CloudRunManager(@Named("K_CONFIGURATION") final String configuration,
      @Named("K_SERVICE") final String service, @Named("K_REVISION") final String revision,
      RuntimeInfoUtils runtimeInfoUtils) {
    this.configuration = configuration;
    this.service = service;
    this.revision = revision;
    this.runtimeInfoUtils = runtimeInfoUtils;
  }


  public MonitoredResource detectCloudRunResource() {
    return MonitoredResource.newBuilder()
        .setType("cloud_run_revision")
        .putLabels("project_id", runtimeInfoUtils.getProjectId())
        .putLabels("location", runtimeInfoUtils.getRegion())
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
    return !runtimeInfoUtils.isNullOrBlank(configuration)
        && !runtimeInfoUtils.isNullOrBlank(service)
        && !runtimeInfoUtils.isNullOrBlank(revision);
  }

}
