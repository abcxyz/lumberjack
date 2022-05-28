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

import com.google.api.client.util.Strings;
import lombok.Data;
import lombok.EqualsAndHashCode;

/**
 * Contains configuration pertaining to RemoteProcessors. Each value defaults to the value in YAML
 * configuration, but may be overridden using environment variables.
 */
@Data
@EqualsAndHashCode
public class BackendContext {
  RemoteConfiguration remote;
  LocalConfiguration local;
  CloudLoggingConfiguration cloudlogging;

  public RemoteConfiguration getRemote() {
    if (remote == null) {
      remote = new RemoteConfiguration();
    }
    return remote;
  }

  public CloudLoggingConfiguration getCloudlogging() {
    if (cloudlogging == null) {
      cloudlogging = new CloudLoggingConfiguration();
      cloudlogging.setDefaultProject(false);
    }
    return cloudlogging;
  }

  public boolean remoteEnabled() {
    return !Strings.isNullOrEmpty(getRemote().getAddress());
  }

  public boolean localLoggingEnabled() {
    return !(local == null) && local.logOutEnabled();
  }

  public boolean cloudLoggingEnabled() {
    // Check that cloud logging config exists. If it does, make sure either project is set or
    // default is enabled.
    return getCloudlogging().useDefaultProject()
        || !Strings.isNullOrEmpty(getCloudlogging().getProject());
  }
}
