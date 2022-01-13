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

package com.abcxyz.lumberjack.auditlogclient.modules;

import com.abcxyz.lumberjack.auditlogclient.config.EnvironmentVariableConfiguration;
import com.google.api.MonitoredResource;
import com.google.cloud.MetadataConfig;
import com.google.inject.AbstractModule;
import com.google.inject.Inject;
import com.google.inject.Provides;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;
import java.io.IOException;
import java.net.URL;
import java.net.URLConnection;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import java.util.Map;
import javax.annotation.Nullable;

/**
 * Provides configuration for runtimeInfo processing.
 */
public class RuntimeInfoProcessorModule extends AbstractModule {

  @Provides
  @Inject
  @Nullable
  public com.google.protobuf.Value monitoredResource(EnvironmentVariableConfiguration envConfig)
      throws IOException {
    MonitoredResource monitoredResource;

    if (isCloudRun(envConfig)) {
      monitoredResource = detectCloudRunResource(envConfig);
    } else if (isAppEngine(envConfig)) {
      monitoredResource = detectAppEngineResource(envConfig);
    } else if (isCloudFunction(envConfig)) {
      monitoredResource = detectCloudFunction(envConfig);
    } else if (isKubernetesEngine()) {
      monitoredResource = detectKubernetesResource(envConfig);
    } else if (isOnGCE(envConfig)) {
      monitoredResource = detectGCEResource(envConfig);
    } else {
      return null;
    }

    com.google.protobuf.Value value = structToVal(monitoredResource);
    return value;
  }

  private MonitoredResource detectGCEResource(EnvironmentVariableConfiguration envConfig) {
    return MonitoredResource.newBuilder()
        .setType("gce_instance")
        .putLabels("project_id", getProjectId())
        .putLabels("instance_id", getInstanceId())
        .putLabels("instance_name", getInstanceName())
        .putLabels("zone", getZone())
        .build();
  }

  private boolean isOnGCE(EnvironmentVariableConfiguration envConfig) throws IOException {
     if (!isNullOrBlank(envConfig.getMetadataHostEnv())) {
       return true;
     }
      URL url = new URL("http://metadata.google.internal");
      URLConnection connection = url.openConnection();
      Map<String, List<String>> map = connection.getHeaderFields();
      List<String> metadataFlavour = map.get("Metadata-Flavor");
      return metadataFlavour.contains("Google");
  }

  private MonitoredResource detectKubernetesResource(EnvironmentVariableConfiguration envConfig)
      throws IOException {
    Path path = Paths.get("/var/run/secrets/kubernetes.io/serviceaccount/namespace");
    byte[] data = Files.readAllBytes(path);
    String namespaceName = new String(data);
    return MonitoredResource.newBuilder()
        .setType("k8s_container")
        .putLabels("project_id", getProjectId())
        .putLabels("cluster_name", getClusterName())
        .putLabels("pod_name", envConfig.getHostname())
        .putLabels("container_name", envConfig.getContainerName())
        .putLabels("namespace_name", namespaceName)
        .putLabels("location", getZone())
        .build();
  }

  private boolean isKubernetesEngine() {
    String clusterName = MetadataConfig.getClusterName();
    return clusterName != null && !clusterName.isBlank();
  }

  private MonitoredResource detectCloudFunction(EnvironmentVariableConfiguration envConfig) {
    String functionName =
        isNullOrBlank(envConfig.getKService())
            ? envConfig.getFunctionName()
            : envConfig.getKService();
    return MonitoredResource.newBuilder()
        .setType("cloud_function")
        .putLabels("project_id", getProjectId())
        .putLabels("function_name", functionName)
        .putLabels("region", getRegion())
        .build();
  }

  private boolean isCloudFunction(EnvironmentVariableConfiguration envConfig) {
    return (!isNullOrBlank(envConfig.getFunctionName())
        && !isNullOrBlank(envConfig.getFunctionRegion())
        && !isNullOrBlank(envConfig.getFunctionPoint()))
        || (!isNullOrBlank(envConfig.getFunctionTarget())
        && !isNullOrBlank(envConfig.getFunctionSignatureType())
        && !isNullOrBlank(envConfig.getKService()));
  }

  private MonitoredResource detectAppEngineResource(EnvironmentVariableConfiguration envConfig) {
    return MonitoredResource.newBuilder()
        .setType("gae_app")
        .putLabels("project_id", getProjectId())
        .putLabels("module_id", envConfig.getGaeService())
        .putLabels("version_id", envConfig.getGaeVersion())
        .putLabels("instance_id", envConfig.getGaeInstance())
        .putLabels("runtime", envConfig.getGaeRuntime())
        .putLabels("zone", getZone())
        .build();
  }

  private boolean isAppEngine(EnvironmentVariableConfiguration envConfig) {
    return !isNullOrBlank(envConfig.getGaeInstance())
        && !isNullOrBlank(envConfig.getGaeService())
        && !isNullOrBlank(envConfig.getGaeVersion());
  }

  private com.google.protobuf.Value structToVal(MonitoredResource monitoredResource)
      throws InvalidProtocolBufferException {
    if (monitoredResource == null) {
      return null;
    }
    String json = JsonFormat.printer().print(monitoredResource);

    com.google.protobuf.Value.Builder responseBuilder = com.google.protobuf.Value.newBuilder();
    JsonFormat.parser().merge(json, responseBuilder);
    return responseBuilder.build();
  }

  private MonitoredResource detectCloudRunResource(EnvironmentVariableConfiguration envConfig) {
    return MonitoredResource.newBuilder()
        .setType("cloud_run_revision")
        .putLabels("project_id", getProjectId())
        .putLabels("location", getRegion())
        .putLabels("service_name", envConfig.getKService())
        .putLabels("revision_name", envConfig.getKRevision())
        .putLabels("configuration_name", envConfig.getKConfiguration())
        .build();
  }

  private String getRegion() {
    String zone = getZone();
    int cutOff = zone.lastIndexOf("-");
    return cutOff > 0 ? zone.substring(0, cutOff) : zone;
  }

  private boolean isCloudRun(EnvironmentVariableConfiguration envConfig) {
    return !isNullOrBlank(envConfig.getKConfiguration())
        && !isNullOrBlank(envConfig.getKService())
        && !isNullOrBlank(envConfig.getKRevision());
  }

  private String getClusterName() {
    String clusterName = MetadataConfig.getClusterName();
    if (clusterName == null || clusterName.isBlank()) {
      throw new IllegalArgumentException("Cluster name not found in metadata.");
    }
    return clusterName;
  }

  private String getZone() {
    String zone = MetadataConfig.getZone();
    if (zone == null || zone.isBlank()) {
      throw new IllegalArgumentException("Zone not found in metadata.");
    }
    return zone;
  }

  private String getProjectId() {
    String projectId = MetadataConfig.getProjectId();
    if (projectId == null || projectId.isBlank()) {
      throw new IllegalArgumentException("Project ID not found in metadata.");
    }
    return projectId;
  }

  private String getInstanceId() {
    String instanceId = MetadataConfig.getInstanceId();
    if (instanceId == null || instanceId.isBlank()) {
      throw new IllegalArgumentException("Instance Id not found in metadata.");
    }
    return instanceId;
  }

  private String getInstanceName() {
    String instanceName = MetadataConfig.getAttribute("instance/name");
    if (instanceName == null || instanceName.isBlank()) {
      throw new IllegalArgumentException("instance Name not found in metadata.");
    }
    return instanceName;
  }

  private boolean isNullOrBlank(String val) {
    return val == null || val.isBlank();
  }

  @Override
  protected void configure() {
  }
}
