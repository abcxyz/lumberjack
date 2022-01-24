package com.abcxyz.lumberjack.auditlogclient.utils;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.name.Named;


public class AppEngineManager {

  private final String service, version, instance, runtime;
  private final RuntimeInfoUtils runtimeInfoUtils;

  @Inject
  public AppEngineManager(@Named("GAE_SERVICE") final String service,
      @Named("GAE_VERSION") final String version, @Named("GAE_INSTANCE") final String instance,
      @Named("GAE_RUNTIME") final String runtime,
      RuntimeInfoUtils runtimeInfoUtils) {
    this.service = service;
    this.version = version;
    this.instance = instance;
    this.runtime = runtime;
    this.runtimeInfoUtils = runtimeInfoUtils;
  }


  public MonitoredResource detectAppEngineResource() {
    return MonitoredResource.newBuilder()
        .setType("gae_app")
        .putLabels("project_id", runtimeInfoUtils.getProjectId())
        .putLabels("module_id", service)
        .putLabels("version_id", version)
        .putLabels("instance_id", instance)
        .putLabels("runtime", runtime)
        .putLabels("zone", runtimeInfoUtils.getZone())
        .build();
  }

  /**
   * Detect if current process is being run on an App Engine. Based on
   * https://cloud.google.com/appengine/docs/standard/java11/runtime#environment_variables
   */
  public boolean isAppEngine() {
    return !runtimeInfoUtils.isNullOrBlank(instance)
        && !runtimeInfoUtils.isNullOrBlank(service)
        && !runtimeInfoUtils.isNullOrBlank(version)
        && !runtimeInfoUtils.isNullOrBlank(runtime);
  }
}
