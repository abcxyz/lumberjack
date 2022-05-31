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
