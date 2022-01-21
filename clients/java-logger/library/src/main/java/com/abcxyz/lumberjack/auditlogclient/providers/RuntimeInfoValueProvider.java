package com.abcxyz.lumberjack.auditlogclient.providers;

import com.google.api.MonitoredResource;
import com.google.cloud.MetadataConfig;
import com.google.inject.Provider;
import com.google.inject.ProvisionException;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Value;
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


public class RuntimeInfoValueProvider implements Provider<Value> {

  private static final String metadataUrl = "http://metadata.google.internal";

  @Override
  @Nullable
  public Value get() {
    MonitoredResource monitoredResource;
    try {
      if (isCloudRun()) {
        monitoredResource = detectCloudRunResource();
      } else if (isAppEngine()) {
        monitoredResource = detectAppEngineResource();
      } else if (isCloudFunction()) {
        monitoredResource = detectCloudFunction();
      } else if (isKubernetesEngine()) {
        monitoredResource = detectKubernetesResource();
      } else if (isOnGCE()) {
        monitoredResource = detectGCEResource();
      } else {
        return null;
      }
      com.google.protobuf.Value value = structToVal(monitoredResource);
      return value;
    } catch (IOException e) {
      throw new ProvisionException("Caught IOException", e);
    }
  }

  protected MonitoredResource detectGCEResource() {
    return MonitoredResource.newBuilder()
        .setType("gce_instance")
        .putLabels("project_id", getProjectId())
        .putLabels("instance_id", getInstanceId())
        .putLabels("instance_hostname", getInstanceName())
        .putLabels("zone", getZone())
        .build();
  }

  public boolean isOnGCE() throws IOException {
    URL url = new URL(metadataUrl);
    URLConnection connection = url.openConnection();
    Map<String, List<String>> map = connection.getHeaderFields();
    List<String> metadataFlavor = map.get("Metadata-Flavor");
    return metadataFlavor.contains("Google");
  }

  private MonitoredResource detectKubernetesResource()
      throws IOException {
    Path path = Paths.get("/var/run/secrets/kubernetes.io/serviceaccount/namespace");
    byte[] data = Files.readAllBytes(path);
    String namespaceName = new String(data);
    String podName = System.getenv().getOrDefault("HOSTNAME", "");
    String containerName = System.getenv().getOrDefault("CONTAINER_NAME", "");
    return MonitoredResource.newBuilder()
        .setType("k8s_container")
        .putLabels("project_id", getProjectId())
        .putLabels("cluster_name", getClusterName())
        .putLabels("pod_name", podName)
        .putLabels("container_name", containerName)
        .putLabels("namespace_name", namespaceName)
        .putLabels("location", getZone())
        .build();
  }

  private boolean isKubernetesEngine() {
    String clusterName = MetadataConfig.getClusterName();
    return clusterName != null && !clusterName.isBlank();
  }

  private MonitoredResource detectCloudFunction() {
    return MonitoredResource.newBuilder()
        .setType("cloud_function")
        .putLabels("project_id", getProjectId())
        .putLabels("function_name", System.getenv("K_SERVICE"))
        .putLabels("region", getRegion())
        .build();
  }

  private boolean isCloudFunction() {
    return
        (!isNullOrBlank(System.getenv("FUNCTION_TARGET"))
            && !isNullOrBlank(System.getenv("FUNCTION_SIGNATURE_TYPE"))
            && !isNullOrBlank(System.getenv("K_SERVICE"))
            && !isNullOrBlank(System.getenv("K_REVISION")));
  }

  private MonitoredResource detectAppEngineResource() {
    return MonitoredResource.newBuilder()
        .setType("gae_app")
        .putLabels("project_id", getProjectId())
        .putLabels("module_id", System.getenv("GAE_SERVICE"))
        .putLabels("version_id", System.getenv("GAE_VERSION"))
        .putLabels("instance_id", System.getenv("GAE_INSTANCE"))
        .putLabels("runtime", System.getenv("GAE_RUNTIME"))
        .putLabels("zone", getZone())
        .build();
  }

  private boolean isAppEngine() {
    return !isNullOrBlank(System.getenv("GAE_INSTANCE"))
        && !isNullOrBlank(System.getenv("GAE_SERVICE"))
        && !isNullOrBlank(System.getenv("GAE_VERSION"))
        && !isNullOrBlank(System.getenv("GAE_RUNTIME"));
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
    return MonitoredResource.newBuilder()
        .setType("cloud_run_revision")
        .putLabels("project_id", getProjectId())
        .putLabels("location", getRegion())
        .putLabels("service_name", System.getenv("K_SERVICE"))
        .putLabels("revision_name", System.getenv("K_REVISION"))
        .putLabels("configuration_name", System.getenv("K_CONFIGURATION"))
        .build();
  }

  private String getRegion() {
    String zone = getZone();
    int cutOff = zone.lastIndexOf("-");
    return cutOff > 0 ? zone.substring(0, cutOff) : zone;
  }

  private boolean isCloudRun() {
    return !isNullOrBlank(System.getenv("K_CONFIGURATION"))
        && !isNullOrBlank(System.getenv("K_SERVICE"))
        && !isNullOrBlank(System.getenv("K_REVISION"));
  }

  private String getClusterName() {
    String clusterName = MetadataConfig.getClusterName();
    if (clusterName == null) {
      throw new IllegalArgumentException("ClusterName returned null from metadata.");
    }
    return clusterName;
  }

  private String getZone() {
    String zone = MetadataConfig.getZone();
    if (zone == null) {
      throw new IllegalArgumentException("Zone not returned null from metadata.");
    }
    return zone;
  }

  private String getProjectId() {
    String projectId = MetadataConfig.getProjectId();
    if (projectId == null) {
      throw new IllegalArgumentException("ProjectID returned null from metadata.");
    }
    return projectId;
  }

  private String getInstanceId() {
    String instanceId = MetadataConfig.getInstanceId();
    if (instanceId == null) {
      throw new IllegalArgumentException("InstanceId returned null from metadata.");
    }
    return instanceId;
  }

  private String getInstanceName() {
    String instanceName = MetadataConfig.getAttribute("instance/name");
    if (instanceName == null) {
      throw new IllegalArgumentException("instanceName returned null from metadata.");
    }
    return instanceName;
  }

  private boolean isNullOrBlank(String val) {
    return val == null || val.isBlank();
  }
}
