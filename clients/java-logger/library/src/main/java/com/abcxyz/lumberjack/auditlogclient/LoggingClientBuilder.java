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

package com.abcxyz.lumberjack.auditlogclient;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.processor.CloudLoggingProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.FilteringProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LabelProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LocalLogProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogBackend;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogMutator;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogValidator;
import com.abcxyz.lumberjack.auditlogclient.processor.RemoteProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.RuntimeInfoProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.ValidationProcessor;
import com.google.inject.Inject;
import java.util.ArrayList;
import java.util.LinkedHashSet;
import lombok.RequiredArgsConstructor;

/** Builder for {@link LoggingClient}. */
@RequiredArgsConstructor(onConstructor = @__({@Inject}))
public class LoggingClientBuilder {
  private final AuditLoggingConfiguration auditLoggingConfiguration;

  private final CloudLoggingProcessor cloudLoggingProcessor;
  private final FilteringProcessor filteringProcessor;
  private final RemoteProcessor remoteProcessor;
  private final RuntimeInfoProcessor runtimeInfoProcessor;
  private final ValidationProcessor validationProcessor;
  private final LabelProcessor labelProcessor;
  private final LocalLogProcessor localLogProcessor;

  private final LinkedHashSet<LogValidator> validators = new LinkedHashSet<>();
  private final LinkedHashSet<LogMutator> mutators = new LinkedHashSet<>();
  private final LinkedHashSet<LogBackend> backends = new LinkedHashSet<>();

  /**
   * Provides a {@link LoggingClientBuilder} with default {@link LogProcessor}s; {@link
   * ValidationProcessor}, {@link RuntimeInfoProcessor}, and {@link RemoteProcessor}. }
   */
  public LoggingClientBuilder withDefaultProcessors() {
    return withValidationProcessor()
        .withFilteringProcessor()
        .withRuntimeInfoProcessor()
        .withLogBackends()
        .withLabelProcessor();
  }

  /** Provides a {@link LoggingClientBuilder} with {@link CloudLoggingProcessor}. */
  public LoggingClientBuilder withCloudLoggingProcessor() {
    backends.add(cloudLoggingProcessor);
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link FilteringProcessor}. */
  public LoggingClientBuilder withFilteringProcessor() {
    mutators.add(filteringProcessor);
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link RuntimeInfoProcessor}. */
  public LoggingClientBuilder withRuntimeInfoProcessor() {
    mutators.add(runtimeInfoProcessor);
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link ValidationProcessor}. */
  public LoggingClientBuilder withValidationProcessor() {
    validators.add(validationProcessor);
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link LogBackend}s. */
  public LoggingClientBuilder withLogBackends() {
    if (auditLoggingConfiguration.getBackend().remoteEnabled()) {
      this.withRemoteProcessor();
    }
    if (auditLoggingConfiguration.getBackend().localLoggingEnabled()) {
      this.withLocalLogProcessor();
    }
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link RemoteProcessor}. */
  public LoggingClientBuilder withRemoteProcessor() {
    backends.add(remoteProcessor);
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link LocalLogProcessor}. */
  public LoggingClientBuilder withLocalLogProcessor() {
    backends.add(localLogProcessor);
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link LabelProcessor}. */
  public LoggingClientBuilder withLabelProcessor() {
    mutators.add(labelProcessor);
    return this;
  }

  /** Constructs and returns the {@link LoggingClient}. */
  public LoggingClient build() {
    return new LoggingClient(
        new ArrayList<>(validators), new ArrayList<>(mutators), new ArrayList<>(backends), auditLoggingConfiguration);
  }
}
