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

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class BackendContextTest {
  @Test
  public void remoteIsEnabled() {
    BackendContext backendContext = new BackendContext();
    backendContext.setAddress("example.com");
    assertThat(backendContext.remoteEnabled()).isTrue();
  }

  @Test
  public void remoteIsDisabled() {
    BackendContext backendContext = new BackendContext();
    assertThat(backendContext.remoteEnabled()).isFalse();
    backendContext.setAddress("");
    assertThat(backendContext.remoteEnabled()).isFalse();
  }
}
