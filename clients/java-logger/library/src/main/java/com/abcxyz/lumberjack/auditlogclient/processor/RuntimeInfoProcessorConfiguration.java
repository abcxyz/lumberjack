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

package com.abcxyz.lumberjack.auditlogclient.processor;

import com.google.api.MonitoredResource;
import com.google.cloud.MetadataConfig;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * Provides configuration for runtimeInfo processing.
 */
@Configuration
public class RuntimeInfoProcessorConfiguration {

  @Value("${K_CONFIGURATION:}")
  private String kConfiguration;

  @Value("${K_SERVICE:}")
  private String kService;

  @Value("${K_REVISION:}")
  private String kRevision;

  @Value("${GOOGLE_CLOUD_PROJECT:}")
  private String googleCloudProjectId;

  @Value("${GAE_SERVICE:}")
  private String gaeService;

  @Value("${GAE_VERSION:}")
  private String gaeVersion;

  @Value("${GAE_INSTANCE:}")
  private String gaeInstance;

  @Value("${GAE_RUNTIME:}")
  private String gaeRuntime;

  @Value("${FUNCTION_NAME:}")
  private String functionName;

  @Value("${FUNCTION_REGION:}")
  private String functionRegion;

  @Value("${ENTRY_POINT:}")
  private String functionPoint;

  @Value("${FUNCTION_TARGET:}")
  private String functionTarget;

  @Value("${FUNCTION_SIGNATURE_TYPE:}")
  private String functionSignatureType;

  @Value("${HOSTNAME:}")
  private String hostname;

  @Value("${CONTAINER_NAME:}")
  private String containerName;

  @Bean
  com.google.protobuf.Value monitoredResource() throws IOException {
    MonitoredResource monitoredResource;

    if (isCloudRun()) {
      monitoredResource = detectCloudRunResource();
    } else if (isAppEngine()) {
      monitoredResource = detectAppEngineResource();
    } else if (isCloudFunction()) {
      monitoredResource = detectCloudFunction();
    } else if (isKubernetesEngine()) {
      monitoredResource = detectKubernetesResource();
    } else {
      return null;
    }
    //TODO(b/205826340): Add GCE runtime info
    com.google.protobuf.Value value = structToVal(monitoredResource);
    return value;
  }

  private MonitoredResource detectKubernetesResource() throws IOException {
    Path path = Paths.get("/var/run/secrets/kubernetes.io/serviceaccount/namespace");
    byte[] data = Files.readAllBytes(path);
    String namespaceName = new String(data);
    return MonitoredResource.newBuilder().setType("k8s_container")
        .putLabels("project_id", getProjectId())
        .putLabels("cluster_name", getClusterName())
        .putLabels("pod_name", hostname)
        .putLabels("container_name", containerName)
        .putLabels("namespace_name", namespaceName)
        .putLabels("location", getZone()).build();
  }

  private boolean isKubernetesEngine() {
    String clusterName = MetadataConfig.getClusterName();
    return clusterName != null && !clusterName.isBlank();
  }


  private MonitoredResource detectCloudFunction() {
    String functionName = kService.isBlank() ? this.functionName : kService;
    return MonitoredResource.newBuilder().setType("cloud_function")
        .putLabels("project_id", getProjectId())
        .putLabels("function_name", functionName)
        .putLabels("region", getRegion()).build();
  }

  private boolean isCloudFunction() {
    return (!functionName.isBlank() && !functionRegion.isBlank() && !functionPoint.isBlank())
        || (!functionTarget.isBlank() && !functionSignatureType.isBlank() && !kService
        .isBlank());
  }

  private MonitoredResource detectAppEngineResource() {
    return MonitoredResource.newBuilder().setType("gae_app")
        .putLabels("project_id", getProjectId())
        .putLabels("module_id", gaeService)
        .putLabels("version_id", gaeVersion)
        .putLabels("instance_id", gaeInstance)
        .putLabels("runtime", gaeRuntime)
        .putLabels("zone", getZone()).build();
  }

  private boolean isAppEngine() {
    return !gaeInstance.isBlank() && !gaeService.isBlank() && !gaeVersion.isBlank();
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

  private MonitoredResource detectCloudRunResource() {
    return MonitoredResource.newBuilder().setType("cloud_run_revision")
        .putLabels("project_id", getProjectId())
        .putLabels("location", getRegion())
        .putLabels("service_name", kService)
        .putLabels("revision_name", kRevision)
        .putLabels("configuration_name", kConfiguration).build();
  }

  private String getRegion() {
    String zone = getZone();
    int cutOff = zone.lastIndexOf("-");
    return cutOff > 0 ? zone.substring(0, cutOff) : zone;
  }

  private boolean isCloudRun() {
    return !kConfiguration.isBlank() && !kService.isBlank() && !kRevision.isBlank();
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
}
