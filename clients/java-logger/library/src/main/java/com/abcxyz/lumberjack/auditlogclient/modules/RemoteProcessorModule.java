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

package com.abcxyz.lumberjack.auditlogclient.modules;

import com.abcxyz.lumberjack.auditlogclient.config.BackendContext;
import com.abcxyz.lumberjack.auditlogclient.config.RemoteConfiguration;
import com.abcxyz.lumberjack.v1alpha1.AuditLogAgentGrpc;
import com.google.auth.oauth2.GoogleCredentials;
import com.google.auth.oauth2.IdTokenCredentials;
import com.google.auth.oauth2.IdTokenProvider;
import com.google.auth.oauth2.IdTokenProvider.Option;
import com.google.auth.oauth2.ImpersonatedCredentials;
import com.google.auth.oauth2.OAuth2Credentials;
import com.google.inject.AbstractModule;
import com.google.inject.Inject;
import com.google.inject.Provides;
import io.grpc.ManagedChannelBuilder;
import io.grpc.auth.MoreCallCredentials;
import java.io.IOException;
import java.util.Collections;
import java.util.List;

public class RemoteProcessorModule extends AbstractModule {
  @Provides
  @Inject
  AuditLogAgentGrpc.AuditLogAgentBlockingStub blockingStub(RemoteConfiguration remoteConfiguration)
      throws IOException {
    if (remoteConfiguration.getInsecureEnabled()) {
      return AuditLogAgentGrpc.newBlockingStub(
          ManagedChannelBuilder.forTarget(remoteConfiguration.getAddress()).usePlaintext().build());
    }

    if (remoteConfiguration.getAddress() == null || remoteConfiguration.getAddress().isBlank()) {
      throw new IllegalArgumentException("AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS must be set.");
    }

    GoogleCredentials credentials = GoogleCredentials.getApplicationDefault();

    if (!(credentials instanceof IdTokenProvider)) {
      throw new IllegalArgumentException(credentials + " not valid as ID token provider.");
    }
    String remoteAudience =
        remoteConfiguration.getAuthAudience() == null || remoteConfiguration.getAuthAudience().isBlank()
            ? "https://" + urlWithoutPort(remoteConfiguration.getAddress())
            : remoteConfiguration.getAuthAudience();

    OAuth2Credentials tokenCredential;

    if (remoteConfiguration.getImpersonateAccount() == null
        || remoteConfiguration.getImpersonateAccount().isBlank()) {
      tokenCredential =
          IdTokenCredentials.newBuilder()
              .setIdTokenProvider((IdTokenProvider) credentials)
              .setTargetAudience(remoteAudience)
              .build();
    } else {
      ImpersonatedCredentials impersonatedCredentials =
          ImpersonatedCredentials.create(
              credentials,
              remoteConfiguration.getImpersonateAccount(),
              /* delegates= */ null,
              /* scopes= */ Collections.emptyList(),
              /* lifetime= */ 0);
      tokenCredential =
          IdTokenCredentials.newBuilder()
              .setIdTokenProvider(impersonatedCredentials)
              .setTargetAudience(remoteAudience)
              .setOptions(List.of(Option.INCLUDE_EMAIL))
              .build();
    }

    return AuditLogAgentGrpc.newBlockingStub(
            ManagedChannelBuilder.forTarget(remoteConfiguration.getAddress()).build())
        .withCallCredentials(MoreCallCredentials.from(tokenCredential));
  }

  private String urlWithoutPort(String url) {
    int idx = url.lastIndexOf(":");
    return idx > 0 ? url.substring(0, idx) : url;
  }

  @Override
  protected void configure() {}
}
