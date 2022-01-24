package com.abcxyz.lumberjack.auditlogclient.Utils;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.auditlogclient.utils.CloudRunManager;
import com.abcxyz.lumberjack.auditlogclient.utils.RuntimeInfoUtils;
import com.google.api.MonitoredResource;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class CloudRunManagerTests {

  @Mock
  RuntimeInfoUtils runtimeInfoUtils;

  @Test
  void WithCorrectEnvironmentVariablesCloudRunReturnsTrue() {
    CloudRunManager cloudRunManager = new CloudRunManager("TestConfig", "TestService",
        "TestRevision", runtimeInfoUtils);
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestConfig");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestService");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestRevision");
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isTrue();
  }

  @Test
  void WithEmptyEnvironmentVariablesCloudRunReturnsFalse() {
    CloudRunManager cloudRunManager = new CloudRunManager("", "TestService", "TestRevision",
        runtimeInfoUtils);
    Mockito.doReturn(true).when(runtimeInfoUtils).isNullOrBlank("");
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isFalse();
  }

  @Test
  void WithNullEnvironmentVariablesCloudRunReturnsFalse() {
    CloudRunManager cloudRunManager = new CloudRunManager(null, "TestService", "TestRevision",
        runtimeInfoUtils);
    Mockito.doReturn(true).when(runtimeInfoUtils).isNullOrBlank(null);
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isFalse();
  }

  @Test
  void detectCloudRunResourceReturnsValidResource() {
    CloudRunManager cloudRunManager = new CloudRunManager("TestConfiguration", "TestService",
        "TestRevision", runtimeInfoUtils);
    Mockito.doReturn("testProject").when(runtimeInfoUtils).getProjectId();
    Mockito.doReturn("testRegion").when(runtimeInfoUtils).getRegion();
    MonitoredResource mr = cloudRunManager.detectCloudRunResource();
    assertThat(mr.containsLabels("location")).isTrue();
    assertThat(mr.containsLabels("service_name")).isTrue();
  }

  @Test
  void detectCloudRunResourceThrowsExceptionOnInValidResource() {
    CloudRunManager cloudRunManager = new CloudRunManager("TestConfiguration", "TestService",
        "TestRevision", runtimeInfoUtils);
    Mockito.doThrow(new IllegalArgumentException("IllegalArgumentException")).when(runtimeInfoUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> cloudRunManager.detectCloudRunResource());
  }
}
