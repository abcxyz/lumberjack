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
public class CloudFunctionManagerTests {

  @Mock RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Test
  void WithCorrectEnvironmentVariablesIsCloudFunctionReturnsTrue() {
    CloudFunctionManager cloudFunctionManager =
        new CloudFunctionManager(
            "TestFunctionSig",
            "FunctionTarget",
            "TestService",
            "TestRevision",
            runtimeInfoCommonUtils);
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestFunctionSig");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("FunctionTarget");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestService");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestRevision");
    Boolean isCloudFunction = cloudFunctionManager.isCloudFunction();
    assertThat(isCloudFunction).isTrue();
  }

  @Test
  void WithEmptyEnvironmentVariablesIsCloudFunctionReturnsFalse() {
    CloudFunctionManager cloudFunctionManager =
        new CloudFunctionManager(
            "FunctionSigType", "", "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank("");
    Boolean isCloudFunction = cloudFunctionManager.isCloudFunction();
    assertThat(isCloudFunction).isFalse();
  }

  @Test
  void WithNullEnvironmentVariablesIsCloudFunctionReturnsFalse() {
    CloudFunctionManager cloudFunctionManager =
        new CloudFunctionManager(
            "FunctionSigType", null, "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank(null);
    Boolean isCloudFunction = cloudFunctionManager.isCloudFunction();
    assertThat(isCloudFunction).isFalse();
  }

  @Test
  void detectCloudFunctionResourceReturnsValidResource() {
    CloudFunctionManager cloudFunctionManager =
        new CloudFunctionManager(
            null, "FunctionTarget", "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn("testProject").when(runtimeInfoCommonUtils).getProjectId();
    Mockito.doReturn("testRegion").when(runtimeInfoCommonUtils).getRegion();
    MonitoredResource mr = cloudFunctionManager.detectCloudFunction();
    assertThat(mr.containsLabels("function_name")).isTrue();
    assertThat(mr.containsLabels("region")).isTrue();
  }

  @Test
  void detectCloudFunctionResourceThrowsExceptionOnInValidResource() {
    CloudFunctionManager cloudFunctionManager =
        new CloudFunctionManager(
            null, "FunctionTarget", "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doThrow(new IllegalArgumentException("IllegalArgumentException"))
        .when(runtimeInfoCommonUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> cloudFunctionManager.detectCloudFunction());
  }
}
