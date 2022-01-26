package com.abcxyz.lumberjack.auditlogclient.utils.runtimeInfo;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.google.api.MonitoredResource;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class CloudRunManagerTests {

  @Mock
  RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Test
  void WithCorrectEnvironmentVariablesCloudRunReturnsTrue() {
    CloudRunManager cloudRunManager = new CloudRunManager("TestConfig", "TestService",
        "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestConfig");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestService");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestRevision");
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isTrue();
  }

  @Test
  void WithEmptyEnvironmentVariablesCloudRunReturnsFalse() {
    CloudRunManager cloudRunManager = new CloudRunManager("", "TestService", "TestRevision",
        runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank("");
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isFalse();
  }

  @Test
  void WithNullEnvironmentVariablesCloudRunReturnsFalse() {
    CloudRunManager cloudRunManager = new CloudRunManager(null, "TestService", "TestRevision",
        runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank(null);
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isFalse();
  }

  @Test
  void detectCloudRunResourceReturnsValidResource() {
    CloudRunManager cloudRunManager = new CloudRunManager("TestConfiguration", "TestService",
        "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn("testProject").when(runtimeInfoCommonUtils).getProjectId();
    Mockito.doReturn("testRegion").when(runtimeInfoCommonUtils).getRegion();
    MonitoredResource mr = cloudRunManager.detectCloudRunResource();
    assertThat(mr.containsLabels("location")).isTrue();
    assertThat(mr.containsLabels("service_name")).isTrue();
  }

  @Test
  void detectCloudRunResourceThrowsExceptionOnInValidResource() {
    CloudRunManager cloudRunManager = new CloudRunManager("TestConfiguration", "TestService",
        "TestRevision", runtimeInfoCommonUtils);
    Mockito.doThrow(new IllegalArgumentException("IllegalArgumentException")).when(
            runtimeInfoCommonUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> cloudRunManager.detectCloudRunResource());
  }
}
