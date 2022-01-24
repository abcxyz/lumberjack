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
import com.google.inject.AbstractModule;
import com.google.inject.name.Names;
import com.google.protobuf.Value;


/**
 * Provides configuration for runtimeInfo processing.
 */
public class RuntimeInfoProcessorModule extends AbstractModule {

  @Override
  protected void configure() {
    bind(Value.class).toProvider(RuntimeInfoValueProvider.class);
    bind(String.class)
        .annotatedWith(Names.named("K_CONFIGURATION"))
        .toInstance(System.getenv().getOrDefault("K_CONFIGURATION", ""));
    bind(String.class)
        .annotatedWith(Names.named("K_SERVICE"))
        .toInstance(System.getenv().getOrDefault("K_SERVICE", ""));
    bind(String.class)
        .annotatedWith(Names.named("K_REVISION"))
        .toInstance(System.getenv().getOrDefault("K_REVISION", ""));
    bind(String.class)
        .annotatedWith(Names.named("GAE_SERVICE"))
        .toInstance(System.getenv().getOrDefault("GAE_SERVICE", ""));
    bind(String.class)
        .annotatedWith(Names.named("GAE_VERSION"))
        .toInstance(System.getenv().getOrDefault("GAE_VERSION", ""));
    bind(String.class)
        .annotatedWith(Names.named("GAE_INSTANCE"))
        .toInstance(System.getenv().getOrDefault("GAE_INSTANCE", ""));
    bind(String.class)
        .annotatedWith(Names.named("GAE_RUNTIME"))
        .toInstance(System.getenv().getOrDefault("GAE_RUNTIME", ""));
    bind(String.class)
        .annotatedWith(Names.named("FUNCTION_SIGNATURE_TYPE"))
        .toInstance(System.getenv().getOrDefault("FUNCTION_SIGNATURE_TYPE", ""));
    bind(String.class)
        .annotatedWith(Names.named("FUNCTION_TARGET"))
        .toInstance(System.getenv().getOrDefault("FUNCTION_TARGET", ""));
    bind(String.class)
        .annotatedWith(Names.named("HOSTNAME"))
        .toInstance(System.getenv().getOrDefault("HOSTNAME", ""));
    bind(String.class)
        .annotatedWith(Names.named("CONTAINER_NAME"))
        .toInstance(System.getenv().getOrDefault("CONTAINER_NAME", ""));
  }
}
