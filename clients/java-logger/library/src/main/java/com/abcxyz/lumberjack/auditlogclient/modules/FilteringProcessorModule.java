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

import com.abcxyz.lumberjack.auditlogclient.config.Filters;
import com.abcxyz.lumberjack.auditlogclient.processor.FilteringProcessor;
import com.google.inject.AbstractModule;
import com.google.inject.Provides;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.regex.Pattern;

public class FilteringProcessorModule extends AbstractModule {
  private List<Pattern> includePatterns(Filters filters) {
    if (filters == null
        || filters.getPrincipalInclude() == null
        || filters.getPrincipalInclude().isBlank()) {
      return Collections.emptyList();
    }
    List<Pattern> includePatternsFromString = new ArrayList<>();
    for (String regex : filters.getPrincipalInclude().split(",")) {
      includePatternsFromString.add(Pattern.compile(regex.strip()));
    }
    return includePatternsFromString;
  }

  private List<Pattern> excludePatterns(Filters filters) {
    if (filters == null
        || filters.getPrincipalExclude() == null
        || filters.getPrincipalExclude().isBlank()) {
      return Collections.emptyList();
    }
    List<Pattern> excludePatternsFromString = new ArrayList<>();
    for (String regex : filters.getPrincipalExclude().split(",")) {
      excludePatternsFromString.add(Pattern.compile(regex.strip()));
    }
    return excludePatternsFromString;
  }

  @Provides
  FilteringProcessor filteringProcessor(Filters filters) {
    return new FilteringProcessor(includePatterns(filters), excludePatterns(filters));
  }

  @Override
  protected void configure() {}
}
