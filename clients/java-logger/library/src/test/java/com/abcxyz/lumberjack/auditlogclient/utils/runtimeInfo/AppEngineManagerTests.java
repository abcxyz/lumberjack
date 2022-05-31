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
public class AppEngineManagerTests {

  @Mock RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Test
  void WithCorrectEnvironmentVariablesIsAppEngineReturnsTrue() {
    AppEngineManager appEngineManager =
        new AppEngineManager(
            "TestService", "TestVersion", "TestInstance", "TestRuntime", runtimeInfoCommonUtils);
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestService");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestVersion");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestInstance");
    Mockito.doReturn(false).when(runtimeInfoCommonUtils).isNullOrBlank("TestRuntime");
    Boolean isAppEngine = appEngineManager.isAppEngine();
    assertThat(isAppEngine).isTrue();
  }

  @Test
  void WithEmptyEnvironmentVariablesIsAppEngineReturnsFalse() {
    AppEngineManager appEngineManager =
        new AppEngineManager(
            "TestService", "TestVersion", "", "TestRuntime", runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank("");
    Boolean isAppEngine = appEngineManager.isAppEngine();
    assertThat(isAppEngine).isFalse();
  }

  @Test
  void WithNullEnvironmentVariablesIsAppEngineReturnsFalse() {
    AppEngineManager appEngineManager =
        new AppEngineManager(
            "TestService", "TestVersion", null, "TestRuntime", runtimeInfoCommonUtils);
    Mockito.doReturn(true).when(runtimeInfoCommonUtils).isNullOrBlank(null);
    Boolean isAppEngine = appEngineManager.isAppEngine();
    assertThat(isAppEngine).isFalse();
  }

  @Test
  void detectAppEngineResourceReturnsValidResource() {
    AppEngineManager appEngineManager =
        new AppEngineManager(
            "TestService", "TestVersion", "TestInstance", "TestRuntime", runtimeInfoCommonUtils);
    Mockito.doReturn("testProject").when(runtimeInfoCommonUtils).getProjectId();
    Mockito.doReturn("testZone").when(runtimeInfoCommonUtils).getZone();
    MonitoredResource mr = appEngineManager.detectAppEngineResource();
    assertThat(mr.containsLabels("project_id")).isTrue();
    assertThat(mr.containsLabels("runtime")).isTrue();
  }

  @Test
  void detectCloudRunResourceThrowsExceptionOnInValidResource() {
    AppEngineManager appEngineManager =
        new AppEngineManager(
            null, "TestVersion", "TestInstance", "TestRuntime", runtimeInfoCommonUtils);
    Mockito.doThrow(new IllegalArgumentException("IllegalArgumentException"))
        .when(runtimeInfoCommonUtils)
        .getProjectId();
    assertThrows(IllegalArgumentException.class, () -> appEngineManager.detectAppEngineResource());
  }
}
