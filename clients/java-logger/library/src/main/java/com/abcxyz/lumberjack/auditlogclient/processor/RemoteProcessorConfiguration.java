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

package com.abcxyz.lumberjack.auditlogclient.processor;

import com.google.auth.oauth2.GoogleCredentials;
import com.google.auth.oauth2.IdTokenCredentials;
import com.google.auth.oauth2.IdTokenProvider;
import com.google.auth.oauth2.IdTokenProvider.Option;
import com.google.auth.oauth2.ImpersonatedCredentials;
import com.google.auth.oauth2.OAuth2Credentials;
import com.abcxyz.lumberjack.v1alpha1.AuditLogAgentGrpc;
import io.grpc.ManagedChannelBuilder;
import io.grpc.auth.MoreCallCredentials;
import java.io.IOException;
import java.util.Collections;
import java.util.List;
import com.abcxyz.lumberjack.auditlogclient.config.YamlFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.PropertySource;

/** Provides configuration for remote logging service processing (via grpc). */
@Configuration
@PropertySource(value = "classpath:application.yml", factory = YamlFactory.class)
public class RemoteProcessorConfiguration {
  @Value("#{'${AUDIT_CLIENT_BACKEND_ADDRESS:${backend.address:}}'}")
  private String backendAddress;

  @Value("#{'${AUDIT_CLIENT_BACKEND_AUTH_AUDIENCE:${backend.auth_audience:}}'}")
  private String backendAuthAudience;

  @Value("#{'${AUDIT_CLIENT_BACKEND_IMPERSONATE_ACCOUNT:${backend.impersonate_account:}}'}")
  private String impersonateAccount;

  // Meant to be true in unit tests only, allows creating `blockingStub` without
  // GoogleCredentials.
  // Setting this to true outside of testing will cause RemoteProcessor calls to
  // fail.
  @Value("#{'${AUDIT_CLIENT_BACKEND_INSECURE_ENABLED:${backend.insecure_enabled:}}'}")
  private boolean backendInsecure;

  @Bean
  AuditLogAgentGrpc.AuditLogAgentBlockingStub blockingStub() throws IOException {
    if (backendInsecure) {
      return AuditLogAgentGrpc.newBlockingStub(
          ManagedChannelBuilder.forTarget(backendAddress).usePlaintext().build());
    }

    if (backendAddress.isBlank()) {
      throw new IllegalArgumentException("AUDIT_CLIENT_BACKEND_ADDRESS must be set.");
    }

    GoogleCredentials credentials = GoogleCredentials.getApplicationDefault();

    if (!(credentials instanceof IdTokenProvider)) {
      throw new IllegalArgumentException(credentials + " not valid as ID token provider.");
    }
    String remoteAudience = backendAuthAudience.isBlank()
        ? "https://" + urlWithoutPort(backendAddress)
        : backendAuthAudience;

    OAuth2Credentials tokenCredential;

    if (impersonateAccount.isBlank()) {
      tokenCredential = IdTokenCredentials.newBuilder()
          .setIdTokenProvider((IdTokenProvider) credentials)
          .setTargetAudience(remoteAudience)
          .build();
    } else {
      ImpersonatedCredentials impersonatedCredentials = ImpersonatedCredentials.create(
          credentials,
          impersonateAccount,
          /* delegates= */ null,
          /* scopes= */ Collections.emptyList(),
          /* lifetime= */ 0);
      tokenCredential = IdTokenCredentials.newBuilder()
          .setIdTokenProvider(impersonatedCredentials)
          .setTargetAudience(remoteAudience)
          .setOptions(List.of(Option.INCLUDE_EMAIL))
          .build();
    }

    return AuditLogAgentGrpc.newBlockingStub(
        ManagedChannelBuilder.forTarget(backendAddress).build())
        .withCallCredentials(MoreCallCredentials.from(tokenCredential));
  }

  private String urlWithoutPort(String url) {
    int idx = url.lastIndexOf(":");
    return idx > 0 ? url.substring(0, idx) : url;
  }
}
