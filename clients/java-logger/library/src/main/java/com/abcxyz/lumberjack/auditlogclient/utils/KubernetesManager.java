package com.abcxyz.lumberjack.auditlogclient.utils;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.name.Named;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

public class KubernetesManager {

  private final String hostname;
  private final String containerName;
  private final RuntimeInfoUtils runtimeInfoUtils;

  @Inject
  public KubernetesManager(@Named("HOSTNAME") final String hostname,
      @Named("CONTAINER_NAME") final String containerName,
      RuntimeInfoUtils runtimeInfoUtils) {
    this.hostname = hostname;
    this.containerName = containerName;
    this.runtimeInfoUtils = runtimeInfoUtils;
  }

  public MonitoredResource detectKubernetesResource() throws IOException {
    return MonitoredResource.newBuilder()
        .setType("k8s_container")
        .putLabels("project_id", runtimeInfoUtils.getProjectId())
        .putLabels("cluster_name", runtimeInfoUtils.getClusterName())
        .putLabels("pod_name", hostname)
        .putLabels("container_name", containerName)
        .putLabels("namespace_name", runtimeInfoUtils.getNamespace())
        .putLabels("location", runtimeInfoUtils.getZone())
        .build();
  }

  public boolean isKubernetesEngine() {
    String clusterName = runtimeInfoUtils.getClusterName();
    return clusterName != null && !clusterName.isBlank();
  }



}
