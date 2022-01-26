package com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo;

import com.google.cloud.MetadataConfig;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;

/**
 * common utilities for runtime info processor
 */
public class RuntimeInfoCommonUtils {

  public String getRegion() {
    String zone = getZone();
    int cutOff = zone.lastIndexOf("-");
    return cutOff > 0 ? zone.substring(0, cutOff) : zone;
  }

  public String getClusterName() {
    String clusterName = MetadataConfig.getClusterName();
    if (clusterName == null) {
      throw new IllegalArgumentException("ClusterName returned null from metadata.");
    }
    return clusterName;
  }

  public String getZone() {
    String zone = MetadataConfig.getZone();
    if (zone == null) {
      throw new IllegalArgumentException("Zone not returned null from metadata.");
    }
    return zone;
  }

  public String getProjectId() {
    String projectId = MetadataConfig.getProjectId();
    if (projectId == null) {
      throw new IllegalArgumentException("ProjectID returned null from metadata.");
    }
    return projectId;
  }

  public String getInstanceId() {
    String instanceId = MetadataConfig.getInstanceId();
    if (instanceId == null) {
      throw new IllegalArgumentException("InstanceId returned null from metadata.");
    }
    return instanceId;
  }

  public String getInstanceName() {
    String instanceName = MetadataConfig.getAttribute("instance/name");
    if (instanceName == null) {
      throw new IllegalArgumentException("instanceName returned null from metadata.");
    }
    return instanceName;
  }

  public boolean isNullOrBlank(String val) {
    return val == null || val.isBlank();
  }

  public String getNamespace() throws IOException {
    Path path = Paths.get("/var/run/secrets/kubernetes.io/serviceaccount/namespace");
    byte[] data = Files.readAllBytes(path);
    return new String(data);
  }
}
