/*
 * Copyright 2021 Lumberjack authors (see AUTHORS file)
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

import com.fasterxml.jackson.annotation.JsonProperty;
import com.google.api.client.util.Strings;
import lombok.Data;

/**
 * Contains configuration pertaining to RemoteProcessors. Each value defaults to the value in YAML
 * configuration, but may be overridden using environment variables.
 */
@Data
public class BackendContext {
  RemoteConfiguration remote;
  LocalConfiguration local;

  public RemoteConfiguration getRemote() {
    if (remote == null) {
      remote = new RemoteConfiguration();
    }
    return remote;
  }

  public boolean remoteEnabled() {
    return !Strings.isNullOrEmpty(getRemote().getAddress());
  }

  public boolean localLoggingEnabled() {
    return !(local == null) && local.logOutEnabled();
  }
}
