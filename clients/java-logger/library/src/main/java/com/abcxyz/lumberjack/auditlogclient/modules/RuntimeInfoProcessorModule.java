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
import com.google.protobuf.Value;

/**
 * Provides configuration for runtimeInfo processing.
 */
public class RuntimeInfoProcessorModule extends AbstractModule {

  @Override
  protected void configure() {
    bind(Value.class).toProvider(RuntimeInfoValueProvider.class);
  }
}
