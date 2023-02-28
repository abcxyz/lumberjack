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
import com.fasterxml.jackson.databind.PropertyNamingStrategies;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import java.io.IOException;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class BackendContextTest {
  private ObjectMapper mapper;

  @BeforeEach
  void setupPropertyNaming() {
    mapper = new ObjectMapper(new YAMLFactory());
    mapper.setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
  }

  @Test
  public void remoteBackend() throws IOException {
    BackendContext backendContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("backend_remote.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isTrue();
    assertThat(backendContext.localLoggingEnabled()).isFalse();
    assertThat(backendContext.cloudLoggingEnabled()).isFalse();
  }

  @Test
  public void localBackend() throws IOException {
    BackendContext backendContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("backend_local.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isFalse();
    assertThat(backendContext.localLoggingEnabled()).isTrue();
    assertThat(backendContext.cloudLoggingEnabled()).isFalse();
  }

  @Test
  public void cloudLoggingBackend_Default() throws IOException {
    BackendContext backendContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("backend_cloudlogging_default.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isFalse();
    assertThat(backendContext.localLoggingEnabled()).isFalse();
    assertThat(backendContext.cloudLoggingEnabled()).isTrue();
  }

  @Test
  public void cloudLoggingBackend_CustomProject() throws IOException {
    BackendContext backendContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("backend_cloudlogging_custom.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isFalse();
    assertThat(backendContext.localLoggingEnabled()).isFalse();
    assertThat(backendContext.cloudLoggingEnabled()).isTrue();
  }

  @Test
  public void allAsBackend() throws IOException {
    BackendContext backendContext =
        mapper
            .readValue(
                getClass().getClassLoader().getResourceAsStream("backend_all.yml"),
                AuditLoggingConfiguration.class)
            .getBackend();

    assertThat(backendContext.remoteEnabled()).isTrue();
    assertThat(backendContext.localLoggingEnabled()).isTrue();
    assertThat(backendContext.cloudLoggingEnabled()).isTrue();
  }
}
