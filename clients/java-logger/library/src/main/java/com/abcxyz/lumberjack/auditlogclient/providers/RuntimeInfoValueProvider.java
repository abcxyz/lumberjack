package com.abcxyz.lumberjack.auditlogclient.providers;

import com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo.AppEngineManager;
import com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo.CloudFunctionManager;
import com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo.CloudRunManager;
import com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo.ComputeEngineManager;
import com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo.KubernetesManager;
import com.google.api.MonitoredResource;
import com.google.inject.Inject;
import com.google.inject.Provider;
import com.google.inject.ProvisionException;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Value;
import com.google.protobuf.util.JsonFormat;
import java.io.IOException;
import javax.annotation.Nullable;

public class RuntimeInfoValueProvider implements Provider<Value> {

  @Inject
  private CloudRunManager cloudRunManager;
  @Inject
  private AppEngineManager appEngineManager;
  @Inject
  private CloudFunctionManager cloudFunctionManager;
  @Inject
  private KubernetesManager kubernetesManager;
  @Inject
  private ComputeEngineManager computeEngineManager;

  @Override
  @Nullable
  public Value get() {
    MonitoredResource monitoredResource;
    try {
      if (cloudRunManager.isCloudRun()) {
        monitoredResource = cloudRunManager.detectCloudRunResource();
      } else if (appEngineManager.isAppEngine()) {
        monitoredResource = appEngineManager.detectAppEngineResource();
      } else if (cloudFunctionManager.isCloudFunction()) {
        monitoredResource = cloudFunctionManager.detectCloudFunction();
      } else if (kubernetesManager.isKubernetesEngine()) {
        monitoredResource = kubernetesManager.detectKubernetesResource();
      } else if (computeEngineManager.isOnGCE()) {
        monitoredResource = computeEngineManager.detectGCEResource();
      } else {
        return null;
      }
      com.google.protobuf.Value value = structToVal(monitoredResource);
      return value;
    } catch (IOException e) {
      throw new ProvisionException("Caught IOException", e);
    }
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
}
