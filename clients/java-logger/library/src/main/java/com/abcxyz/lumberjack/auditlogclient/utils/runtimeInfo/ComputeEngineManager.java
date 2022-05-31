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
import java.io.IOException;
import java.net.URL;
import java.net.URLConnection;
import java.util.List;
import java.util.Map;

/**
 * ComputeEngineManager provides functionality for getting run time info for processes running on
 * GCE.
 */
public class ComputeEngineManager {

  private static final String metadataUrl = "http://metadata.google.internal";

  @Inject private RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  public MonitoredResource detectGCEResource() {
    return MonitoredResource.newBuilder()
        .setType("gce_instance")
        .putLabels("project_id", runtimeInfoCommonUtils.getProjectId())
        .putLabels("instance_id", runtimeInfoCommonUtils.getInstanceId())
        .putLabels("instance_name", runtimeInfoCommonUtils.getInstanceName())
        .putLabels("zone", runtimeInfoCommonUtils.getZone())
        .build();
  }

  public boolean isOnGCE() throws IOException {
    URL url = new URL(metadataUrl);
    URLConnection connection = url.openConnection();
    Map<String, List<String>> map = connection.getHeaderFields();
    List<String> metadataFlavor = map.get("Metadata-Flavor");
    return metadataFlavor != null && metadataFlavor.contains("Google");
  }
}
