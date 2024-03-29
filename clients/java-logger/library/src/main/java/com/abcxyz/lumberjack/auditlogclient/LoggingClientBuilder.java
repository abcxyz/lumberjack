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
import com.abcxyz.lumberjack.auditlogclient.config.Justification;
import com.abcxyz.lumberjack.auditlogclient.processor.CloudLoggingProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.FilteringProcessor;
import com.abcxyz.lumberjack.auditlogclient.processor.JustificationProcessor;
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
import com.google.inject.Injector;
import java.util.ArrayList;
import java.util.LinkedHashSet;
import lombok.Getter;
import lombok.RequiredArgsConstructor;

/** Builder for {@link LoggingClient}. */
@RequiredArgsConstructor(onConstructor = @__({@Inject}))
public class LoggingClientBuilder {
  private final AuditLoggingConfiguration auditLoggingConfiguration;
  private final Injector injector;

  @Getter(lazy = true)
  private final CloudLoggingProcessor cloudLoggingProcessor =
      getProcessor(CloudLoggingProcessor.class);

  @Getter(lazy = true)
  private final FilteringProcessor filteringProcessor = getProcessor(FilteringProcessor.class);

  @Getter(lazy = true)
  private final RemoteProcessor remoteProcessor = getProcessor(RemoteProcessor.class);

  @Getter(lazy = true)
  private final RuntimeInfoProcessor runtimeInfoProcessor =
      getProcessor(RuntimeInfoProcessor.class);

  @Getter(lazy = true)
  private final ValidationProcessor validationProcessor = getProcessor(ValidationProcessor.class);

  @Getter(lazy = true)
  private final LabelProcessor labelProcessor = getProcessor(LabelProcessor.class);

  @Getter(lazy = true)
  private final LocalLogProcessor localLogProcessor = getProcessor(LocalLogProcessor.class);

  @Getter(lazy = true)
  private final JustificationProcessor justificationProcessor =
      getProcessor(JustificationProcessor.class);

  private final LinkedHashSet<LogValidator> validators = new LinkedHashSet<>();
  private final LinkedHashSet<LogMutator> mutators = new LinkedHashSet<>();
  private final LinkedHashSet<LogBackend> backends = new LinkedHashSet<>();

  private <T> T getProcessor(Class<T> clazz) {
    return injector.getInstance(clazz);
  }

  /**
   * Provides a {@link LoggingClientBuilder} with default {@link LogProcessor}s; {@link
   * ValidationProcessor}, {@link RuntimeInfoProcessor}, {@link RemoteProcessor}, and when enabled,
   * {@link JustificationProcessor}. }
   */
  public LoggingClientBuilder withDefaultProcessors() {
    return withValidationProcessor()
        .withFilteringProcessor()
        .withRuntimeInfoProcessor()
        .withDefaultLogBackends()
        .withLabelProcessor()
        .withJustificationProcessorIfEnabled();
  }

  /** Provides a {@link LoggingClientBuilder} with {@link CloudLoggingProcessor}. */
  public LoggingClientBuilder withCloudLoggingProcessor() {
    backends.add(getCloudLoggingProcessor());
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link FilteringProcessor}. */
  public LoggingClientBuilder withFilteringProcessor() {
    validators.add(getFilteringProcessor());
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link RuntimeInfoProcessor}. */
  public LoggingClientBuilder withRuntimeInfoProcessor() {
    mutators.add(getRuntimeInfoProcessor());
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link JustificationProcessor} if enabled. */
  public LoggingClientBuilder withJustificationProcessorIfEnabled() {
    final Justification justification = auditLoggingConfiguration.getJustification();
    if (justification != null && justification.isEnabled()) {
      mutators.add(getJustificationProcessor());
    }
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link ValidationProcessor}. */
  public LoggingClientBuilder withValidationProcessor() {
    validators.add(getValidationProcessor());
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link LogBackend}s. */
  public LoggingClientBuilder withDefaultLogBackends() {
    if (auditLoggingConfiguration.getBackend().remoteEnabled()) {
      this.withRemoteProcessor();
    }
    if (auditLoggingConfiguration.getBackend().localLoggingEnabled()) {
      this.withLocalLogProcessor();
    }
    if (auditLoggingConfiguration.getBackend().cloudLoggingEnabled()) {
      this.withCloudLoggingProcessor();
    }
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link RemoteProcessor}. */
  public LoggingClientBuilder withRemoteProcessor() {
    backends.add(getRemoteProcessor());
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link LocalLogProcessor}. */
  public LoggingClientBuilder withLocalLogProcessor() {
    backends.add(getLocalLogProcessor());
    return this;
  }

  /** Provides a {@link LoggingClientBuilder} with {@link LabelProcessor}. */
  public LoggingClientBuilder withLabelProcessor() {
    mutators.add(getLabelProcessor());
    return this;
  }

  /** Constructs and returns the {@link LoggingClient}. */
  public LoggingClient build() {
    return new LoggingClient(
        new ArrayList<>(validators),
        new ArrayList<>(mutators),
        new ArrayList<>(backends),
        auditLoggingConfiguration);
  }
}
