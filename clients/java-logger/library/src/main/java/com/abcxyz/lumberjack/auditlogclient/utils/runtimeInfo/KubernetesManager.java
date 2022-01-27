package com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.name.Named;
import java.io.IOException;

/**
 * KubernetesManager provides functionality for getting run time info for processes running on GKE.
 */
public class KubernetesManager {

  private final String hostname;
  private final String containerName;
  private final RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Inject
  public KubernetesManager(
      @Named("HOSTNAME") final String hostname,
      @Named("CONTAINER_NAME") final String containerName,
      RuntimeInfoCommonUtils runtimeInfoCommonUtils) {
    this.hostname = hostname;
    this.containerName = containerName;
    this.runtimeInfoCommonUtils = runtimeInfoCommonUtils;
  }

  public MonitoredResource detectKubernetesResource() throws IOException {
    return MonitoredResource.newBuilder()
        .setType("k8s_container")
        .putLabels("project_id", runtimeInfoCommonUtils.getProjectId())
        .putLabels("cluster_name", runtimeInfoCommonUtils.getClusterName())
        .putLabels("pod_name", hostname)
        .putLabels("container_name", containerName)
        .putLabels("namespace_name", runtimeInfoCommonUtils.getNamespace())
        .putLabels("location", runtimeInfoCommonUtils.getZone())
        .build();
  }

  public boolean isKubernetesEngine() {
    return runtimeInfoCommonUtils.hasClusterName();
  }
}
