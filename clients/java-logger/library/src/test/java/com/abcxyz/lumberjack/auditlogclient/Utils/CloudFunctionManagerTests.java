package com.abcxyz.lumberjack.auditlogclient.Utils;


import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.auditlogclient.utils.CloudFunctionManager;
import com.abcxyz.lumberjack.auditlogclient.utils.RuntimeInfoUtils;
import com.google.api.MonitoredResource;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class CloudFunctionManagerTests {

  @Mock
  RuntimeInfoUtils runtimeInfoUtils;

  @Test
  void WithCorrectEnvironmentVariablesIsCloudFunctionReturnsTrue() {
    CloudFunctionManager cloudFunctionManager = new CloudFunctionManager("TestFunctionSig",
        "FunctionTarget", "TestService", "TestRevision", runtimeInfoUtils);
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestFunctionSig");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("FunctionTarget");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestService");
    Mockito.doReturn(false).when(runtimeInfoUtils).isNullOrBlank("TestRevision");
    Boolean isCloudFunction = cloudFunctionManager.isCloudFunction();
    assertThat(isCloudFunction).isTrue();
  }

  @Test
  void WithEmptyEnvironmentVariablesIsCloudFunctionReturnsFalse() {
    CloudFunctionManager cloudFunctionManager = new CloudFunctionManager("FunctionSigType",
        "", "TestService", "TestRevision", runtimeInfoUtils);
    Mockito.doReturn(true).when(runtimeInfoUtils).isNullOrBlank("");
    Boolean isCloudFunction = cloudFunctionManager.isCloudFunction();
    assertThat(isCloudFunction).isFalse();
  }

  @Test
  void WithNullEnvironmentVariablesIsCloudFunctionReturnsFalse() {
    CloudFunctionManager cloudFunctionManager = new CloudFunctionManager("FunctionSigType",
        null, "TestService", "TestRevision", runtimeInfoUtils);
    Mockito.doReturn(true).when(runtimeInfoUtils).isNullOrBlank(null);
    Boolean isCloudFunction = cloudFunctionManager.isCloudFunction();
    assertThat(isCloudFunction).isFalse();
  }

  @Test
  void detectCloudFunctionResourceReturnsValidResource() {
    CloudFunctionManager cloudFunctionManager = new CloudFunctionManager(null,
        "FunctionTarget", "TestService", "TestRevision", runtimeInfoUtils);
    Mockito.doReturn("testProject").when(runtimeInfoUtils).getProjectId();
    Mockito.doReturn("testRegion").when(runtimeInfoUtils).getRegion();
    MonitoredResource mr = cloudFunctionManager.detectCloudFunction();
    assertThat(mr.containsLabels("function_name")).isTrue();
    assertThat(mr.containsLabels("region")).isTrue();
  }

  @Test
  void detectCloudFunctionResourceThrowsExceptionOnInValidResource() {
    CloudFunctionManager cloudFunctionManager = new CloudFunctionManager(null,
        "FunctionTarget", "TestService", "TestRevision", runtimeInfoUtils);
    Mockito.doThrow(new IllegalArgumentException("IllegalArgumentException")).when(runtimeInfoUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> cloudFunctionManager.detectCloudFunction());
  }
}
