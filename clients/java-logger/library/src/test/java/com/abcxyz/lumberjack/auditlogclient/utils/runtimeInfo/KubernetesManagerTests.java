package com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.google.api.MonitoredResource;
import java.io.IOException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class KubernetesManagerTests {

  @Mock
  RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Test
  void detectGCEResourceReturnsResourceOnValidMetadata() throws IOException {
    KubernetesManager kubernetesManager = new KubernetesManager("hostname", "containerName",
        runtimeInfoCommonUtils);
    Mockito.doReturn("testProject").when(runtimeInfoCommonUtils).getProjectId();
    Mockito.doReturn("testClusterName").when(runtimeInfoCommonUtils).getClusterName();
    Mockito.doReturn("testZone").when(runtimeInfoCommonUtils).getZone();
    Mockito.doReturn("TestNamespace").when(runtimeInfoCommonUtils).getNamespace();
    MonitoredResource mr = kubernetesManager.detectKubernetesResource();
    assertThat(mr.containsLabels("project_id")).isTrue();
    assertThat(mr.containsLabels("pod_name")).isTrue();
  }


  void detectGCEResourceThrowsExceptionResourceOnInvalidMetadata() {
    KubernetesManager kubernetesManager = new KubernetesManager("hostname", "containerName",
        runtimeInfoCommonUtils);
    Mockito.doThrow(new IllegalArgumentException("Exception")).when(runtimeInfoCommonUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class,
        () -> kubernetesManager.detectKubernetesResource());
  }

}
