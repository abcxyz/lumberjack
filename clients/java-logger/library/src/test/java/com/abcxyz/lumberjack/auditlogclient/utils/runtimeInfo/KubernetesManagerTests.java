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
import java.io.IOException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class KubernetesManagerTests {

  @Mock RuntimeInfoCommonUtils runtimeInfoCommonUtils;

  @Test
  void detectGCEResourceReturnsResourceOnValidMetadata() throws IOException {
    KubernetesManager kubernetesManager =
        new KubernetesManager("hostname", "containerName", runtimeInfoCommonUtils);
    Mockito.doReturn("testProject").when(runtimeInfoCommonUtils).getProjectId();
    Mockito.doReturn("testClusterName").when(runtimeInfoCommonUtils).getClusterName();
    Mockito.doReturn("testZone").when(runtimeInfoCommonUtils).getZone();
    Mockito.doReturn("TestNamespace").when(runtimeInfoCommonUtils).getNamespace();
    MonitoredResource mr = kubernetesManager.detectKubernetesResource();
    assertThat(mr.containsLabels("project_id")).isTrue();
    assertThat(mr.containsLabels("pod_name")).isTrue();
  }

  void detectGCEResourceThrowsExceptionResourceOnInvalidMetadata() {
    KubernetesManager kubernetesManager =
        new KubernetesManager("hostname", "containerName", runtimeInfoCommonUtils);
    Mockito.doThrow(new IllegalArgumentException("Exception"))
        .when(runtimeInfoCommonUtils)
        .getProjectId();
    assertThrows(
        IllegalArgumentException.class, () -> kubernetesManager.detectKubernetesResource());
  }
}
