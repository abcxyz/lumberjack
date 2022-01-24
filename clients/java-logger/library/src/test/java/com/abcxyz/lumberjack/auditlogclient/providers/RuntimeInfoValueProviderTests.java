package com.abcxyz.lumberjack.auditlogclient.providers;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.auditlogclient.utils.AppEngineManager;
import com.abcxyz.lumberjack.auditlogclient.utils.CloudFunctionManager;
import com.abcxyz.lumberjack.auditlogclient.utils.CloudRunManager;
import com.abcxyz.lumberjack.auditlogclient.utils.ComputeEngineManager;
import com.abcxyz.lumberjack.auditlogclient.utils.KubernetesManager;
import com.google.api.MonitoredResource;
import com.google.inject.ProvisionException;
import com.google.protobuf.Value;
import java.io.IOException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class RuntimeInfoValueProviderTests {

  MonitoredResource mr = MonitoredResource.newBuilder()
      .setType("gce_instance")
      .putLabels("project_id", "gcp_project")
      .putLabels("instance_id", "testInstance")
      .putLabels("instance_hostname", "testInstanceName")
      .putLabels("zone", "testZone")
      .build();

  @Mock
  private CloudRunManager cloudRunManager;
  @Mock
  private CloudFunctionManager cloudFunctionManager;
  @Mock
  private KubernetesManager kubernetesManager;
  @Mock
  private AppEngineManager appEngineManager;
  @Mock
  private ComputeEngineManager computeEngineManager;

  @InjectMocks
  private RuntimeInfoValueProvider runtimeInfoValueProvider;

  @Test
  void getValueWithNoPlatformMatchReturnsNull() throws IOException {
    Mockito.doReturn(false).when(cloudRunManager).isCloudRun();
    Mockito.doReturn(false).when(appEngineManager).isAppEngine();
    Mockito.doReturn(false).when(cloudFunctionManager).isCloudFunction();
    Mockito.doReturn(false).when(kubernetesManager).isKubernetesEngine();
    Mockito.doReturn(false).when(computeEngineManager).isOnGCE();
    Value value = runtimeInfoValueProvider.get();
    assertThat(value).isNull();
  }

  @Test
  void getValueWithGCEPlatformMatchReturns() throws IOException {
    Mockito.doReturn(false).when(cloudRunManager).isCloudRun();
    Mockito.doReturn(false).when(appEngineManager).isAppEngine();
    Mockito.doReturn(false).when(cloudFunctionManager).isCloudFunction();
    Mockito.doReturn(false).when(kubernetesManager).isKubernetesEngine();
    Mockito.doReturn(true).when(computeEngineManager).isOnGCE();
    Mockito.doReturn(mr).when(computeEngineManager).detectGCEResource();
    Value value = runtimeInfoValueProvider.get();
    assertThat(value).isNotNull();
    assertThat(value.getStructValue().containsFields("type")).isEqualTo(true);
    assertThat(value.getStructValue().containsFields("labels")).isEqualTo(true);
  }

  @Test
  void getValueWithIOExceptionFromIsOnGCEThrowsException() throws IOException {
    Mockito.doReturn(false).when(cloudRunManager).isCloudRun();
    Mockito.doReturn(false).when(appEngineManager).isAppEngine();
    Mockito.doReturn(false).when(cloudFunctionManager).isCloudFunction();
    Mockito.doReturn(false).when(kubernetesManager).isKubernetesEngine();
    Mockito.doThrow(new IOException("IOException")).when(computeEngineManager).isOnGCE();
    assertThrows(ProvisionException.class, () -> runtimeInfoValueProvider.get());
  }
}
