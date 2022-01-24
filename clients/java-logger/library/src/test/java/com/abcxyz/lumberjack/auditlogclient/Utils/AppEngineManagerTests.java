package com.abcxyz.lumberjack.auditlogclient.Utils;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.auditlogclient.utils.AppEngineManager;
import com.abcxyz.lumberjack.auditlogclient.utils.RuntimeInfoUtils;
import com.google.api.MonitoredResource;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class AppEngineManagerTests {

  @Mock
  RuntimeInfoUtils runtimeInfoUtils;

  @Test
  void WithCorrectEnvironmentVariablesIsAppEngineReturnsTrue() {
    AppEngineManager appEngineManager = new AppEngineManager("TestService", "TestVersion",
        "TestInstance", "TestRuntime", runtimeInfoUtils);
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestService");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestVersion");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestInstance");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestRuntime");
    Boolean isAppEngine = appEngineManager.isAppEngine();
    assertThat(isAppEngine).isTrue();
  }

  @Test
  void WithEmptyEnvironmentVariablesIsAppEngineReturnsFalse() {
    AppEngineManager appEngineManager = new AppEngineManager("TestService", "TestVersion",
        "", "TestRuntime", runtimeInfoUtils);
    Mockito.doReturn(true).when(runtimeInfoUtils).isNullOrBlank("");
    Boolean isAppEngine = appEngineManager.isAppEngine();
    assertThat(isAppEngine).isFalse();
  }

  @Test
  void WithNullEnvironmentVariablesIsAppEngineReturnsFalse() {
    AppEngineManager appEngineManager = new AppEngineManager("TestService", "TestVersion",
        null, "TestRuntime", runtimeInfoUtils);
    Mockito.doReturn(true).when(runtimeInfoUtils).isNullOrBlank(null);
    Boolean isAppEngine = appEngineManager.isAppEngine();
    assertThat(isAppEngine).isFalse();
  }

  @Test
  void detectAppEngineResourceReturnsValidResource() {
    AppEngineManager appEngineManager = new AppEngineManager("TestService", "TestVersion",
        "TestInstance", "TestRuntime", runtimeInfoUtils);
    Mockito.doReturn("testProject").when(runtimeInfoUtils).getProjectId();
    Mockito.doReturn("testZone").when(runtimeInfoUtils).getZone();
    MonitoredResource mr = appEngineManager.detectAppEngineResource();
    assertThat(mr.containsLabels("project_id")).isTrue();
    assertThat(mr.containsLabels("runtime")).isTrue();
  }

  @Test
  void detectCloudRunResourceThrowsExceptionOnInValidResource() {
    AppEngineManager appEngineManager = new AppEngineManager(null, "TestVersion",
        "TestInstance", "TestRuntime", runtimeInfoUtils);
    Mockito.doThrow(new IllegalArgumentException("IllegalArgumentException")).when(runtimeInfoUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> appEngineManager.detectAppEngineResource());
  }
}
