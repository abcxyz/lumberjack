/*
 * Copyright 2022 Lumberjack authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

  @Mock RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Test
  void WithCorrectEnvironmentVariablesCloudRunReturnsTrue() {
    CloudRunManager cloudRunManager =
        new CloudRunManager("TestConfig", "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestConfig");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestService");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestRevision");
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isTrue();
  }

  @Test
  void WithEmptyEnvironmentVariablesCloudRunReturnsFalse() {
    CloudRunManager cloudRunManager =
        new CloudRunManager("", "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank("");
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isFalse();
  }

  @Test
  void WithNullEnvironmentVariablesCloudRunReturnsFalse() {
    CloudRunManager cloudRunManager =
        new CloudRunManager(null, "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank(null);
    Boolean isCloudRun = cloudRunManager.isCloudRun();
    assertThat(isCloudRun).isFalse();
  }

  @Test
  void detectCloudRunResourceReturnsValidResource() {
    CloudRunManager cloudRunManager =
        new CloudRunManager(
            "TestConfiguration", "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doReturn("testProject").when(runtimeInfoCommonUtils).getProjectId();
    Mockito.doReturn("testRegion").when(runtimeInfoCommonUtils).getRegion();
    MonitoredResource mr = cloudRunManager.detectCloudRunResource();
    assertThat(mr.containsLabels("location")).isTrue();
    assertThat(mr.containsLabels("service_name")).isTrue();
  }

  @Test
  void detectCloudRunResourceThrowsExceptionOnInValidResource() {
    CloudRunManager cloudRunManager =
        new CloudRunManager(
            "TestConfiguration", "TestService", "TestRevision", runtimeInfoCommonUtils);
    Mockito.doThrow(new IllegalArgumentException("IllegalArgumentException"))
        .when(runtimeInfoCommonUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> cloudRunManager.detectCloudRunResource());
  }
}
