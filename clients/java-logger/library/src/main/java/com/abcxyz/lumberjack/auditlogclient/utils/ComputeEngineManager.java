package com.abcxyz.lumberjack.auditlogclient.utils;

import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import java.io.IOException;
import java.net.URL;
import java.net.URLConnection;
import java.util.List;
import java.util.Map;

public class ComputeEngineManager {

  private static final String metadataUrl = "http://metadata.google.internal";

  @Inject
  private RuntimeInfoUtils runtimeInfoUtils ;

  public MonitoredResource detectGCEResource() {
    return MonitoredResource.newBuilder()
        .setType("gce_instance")
        .putLabels("project_id", runtimeInfoUtils.getProjectId())
        .putLabels("instance_id", runtimeInfoUtils.getInstanceId())
        .putLabels("instance_name", runtimeInfoUtils.getInstanceName())
        .putLabels("zone", runtimeInfoUtils.getZone())
        .build();
  }

  public boolean isOnGCE() throws IOException {
    URL url = new URL(metadataUrl);
    URLConnection connection = url.openConnection();
    Map<String, List<String>> map = connection.getHeaderFields();
    List<String> metadataFlavor = map.get("Metadata-Flavor");
    return metadataFlavor.contains("Google");
  }

}
