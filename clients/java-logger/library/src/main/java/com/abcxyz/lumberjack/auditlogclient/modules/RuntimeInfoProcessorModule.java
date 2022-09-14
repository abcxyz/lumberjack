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

import com.abcxyz.lumberjack.auditlogclient.providers.RuntimeInfoValueProvider;
import com.abcxyz.lumberjack.auditlogclient.utils.ConfigUtils;
import com.google.inject.AbstractModule;
import com.google.inject.name.Names;
import com.google.protobuf.Value;

/** Provides configuration for runtimeInfo processing. */
public class RuntimeInfoProcessorModule extends AbstractModule {

  @Override
  protected void configure() {
    // bind provider class
    bind(Value.class).toProvider(RuntimeInfoValueProvider.class);

    // Environment variables bindings for cloud run/function
    bind(String.class)
        .annotatedWith(Names.named("K_CONFIGURATION"))
        .toInstance(ConfigUtils.getEnvOrDefault("K_CONFIGURATION", ""));
    bind(String.class)
        .annotatedWith(Names.named("K_SERVICE"))
        .toInstance(ConfigUtils.getEnvOrDefault("K_SERVICE", ""));
    bind(String.class)
        .annotatedWith(Names.named("K_REVISION"))
        .toInstance(ConfigUtils.getEnvOrDefault("K_REVISION", ""));

    // Environment variables bindings for App Engine
    bind(String.class)
        .annotatedWith(Names.named("GAE_SERVICE"))
        .toInstance(ConfigUtils.getEnvOrDefault("GAE_SERVICE", ""));
    bind(String.class)
        .annotatedWith(Names.named("GAE_VERSION"))
        .toInstance(ConfigUtils.getEnvOrDefault("GAE_VERSION", ""));
    bind(String.class)
        .annotatedWith(Names.named("GAE_INSTANCE"))
        .toInstance(ConfigUtils.getEnvOrDefault("GAE_INSTANCE", ""));
    bind(String.class)
        .annotatedWith(Names.named("GAE_RUNTIME"))
        .toInstance(ConfigUtils.getEnvOrDefault("GAE_RUNTIME", ""));

    // Environment variables bindings for cloud function
    bind(String.class)
        .annotatedWith(Names.named("FUNCTION_SIGNATURE_TYPE"))
        .toInstance(ConfigUtils.getEnvOrDefault("FUNCTION_SIGNATURE_TYPE", ""));
    bind(String.class)
        .annotatedWith(Names.named("FUNCTION_TARGET"))
        .toInstance(ConfigUtils.getEnvOrDefault("FUNCTION_TARGET", ""));

    // Environment variables bindings for k8s
    bind(String.class)
        .annotatedWith(Names.named("HOSTNAME"))
        .toInstance(ConfigUtils.getEnvOrDefault("HOSTNAME", ""));
    bind(String.class)
        .annotatedWith(Names.named("CONTAINER_NAME"))
        .toInstance(ConfigUtils.getEnvOrDefault("CONTAINER_NAME", ""));
  }
}
