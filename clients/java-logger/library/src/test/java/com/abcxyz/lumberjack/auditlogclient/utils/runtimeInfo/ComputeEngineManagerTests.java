package com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.google.api.MonitoredResource;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class ComputeEngineManagerTests {

  @Mock RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @InjectMocks ComputeEngineManager computeEngineManager;

  @Test
  void detectGCEResourceReturnsResourceOnValidMetadata() {
    Mockito.doReturn("testProject").when(runtimeInfoCommonUtils).getProjectId();
    Mockito.doReturn("testInstanceId").when(runtimeInfoCommonUtils).getInstanceId();
    Mockito.doReturn("testInstanceName").when(runtimeInfoCommonUtils).getInstanceName();
    Mockito.doReturn("testZone").when(runtimeInfoCommonUtils).getZone();
    MonitoredResource mr = computeEngineManager.detectGCEResource();
    assertThat(mr.containsLabels("project_id")).isTrue();
    assertThat(mr.containsLabels("instance_name")).isTrue();
  }

  @Test
  void detectGCEResourceThrowsExceptionResourceOnInvalidMetadata() {
    Mockito.doThrow(new IllegalArgumentException("Exception"))
        .when(runtimeInfoCommonUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> computeEngineManager.detectGCEResource());
  }
}
