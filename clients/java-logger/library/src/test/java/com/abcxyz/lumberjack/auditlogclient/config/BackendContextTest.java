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

package com.abcxyz.lumberjack.auditlogclient.config;

import static org.assertj.core.api.Assertions.assertThat;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import java.io.IOException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class BackendContextTest {
  @Test
  public void remoteBackend() throws IOException {
    ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
    BackendContext backendContext=
        mapper
            .readValue(
                this.getClass().getClassLoader().getResourceAsStream("backend_remote.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isTrue();
    assertThat(backendContext.localLoggingEnabled()).isFalse();
  }

  @Test
  public void localBackend() throws IOException {
    ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
    BackendContext backendContext=
        mapper
            .readValue(
                this.getClass().getClassLoader().getResourceAsStream("backend_local.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isFalse();
    assertThat(backendContext.localLoggingEnabled()).isTrue();
  }

  @Test
  public void bothAsBackend() throws IOException {
    ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
    BackendContext backendContext=
        mapper
            .readValue(
                this.getClass().getClassLoader().getResourceAsStream("backend_both.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isTrue();
    assertThat(backendContext.localLoggingEnabled()).isTrue();
  }
}
