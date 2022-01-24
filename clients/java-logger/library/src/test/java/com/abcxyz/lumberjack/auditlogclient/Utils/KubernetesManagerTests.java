package com.abcxyz.lumberjack.auditlogclient.Utils;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.auditlogclient.utils.KubernetesManager;
import com.abcxyz.lumberjack.auditlogclient.utils.RuntimeInfoUtils;
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
  RuntimeInfoUtils runtimeInfoUtils;

  @Test
  void detectGCEResourceReturnsResourceOnValidMetadata() throws IOException {
    KubernetesManager kubernetesManager = new KubernetesManager("hostname", "containerName",
        runtimeInfoUtils);
    Mockito.doReturn("testProject").when(runtimeInfoUtils).getProjectId();
    Mockito.doReturn("testClusterName").when(runtimeInfoUtils).getClusterName();
    Mockito.doReturn("testZone").when(runtimeInfoUtils).getZone();
    Mockito.doReturn("TestNamespace").when(runtimeInfoUtils).getNamespace();
    MonitoredResource mr = kubernetesManager.detectKubernetesResource();
    assertThat(mr.containsLabels("project_id")).isTrue();
    assertThat(mr.containsLabels("pod_name")).isTrue();
  }


  void detectGCEResourceThrowsExceptionResourceOnInvalidMetadata() {
    KubernetesManager kubernetesManager = new KubernetesManager("hostname", "containerName",
        runtimeInfoUtils);
    Mockito.doThrow(new IllegalArgumentException("Exception")).when(runtimeInfoUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class,
        () -> kubernetesManager.detectKubernetesResource());
  }

}
