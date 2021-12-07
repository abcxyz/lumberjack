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

import java.util.ArrayList;
import java.util.List;
import java.util.regex.Pattern;
import com.abcxyz.lumberjack.auditlogclient.config.YamlFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.PropertySource;

/** Provides configuration for {@link FilteringProcessor} */
@Configuration
@PropertySource(value = "classpath:application.yml", factory = YamlFactory.class)
public class FilteringProcessorConfiguration {
  @Value("#{'${AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_INCLUDE:${filter.regex.principal_include:}}'}")
  private String includes;

  @Value("#{'${AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_EXCLUDE:${filter.regex.principal_exclude:}}'}")
  private String excludes;

  @Bean
  List<Pattern> includePatterns() {
    List<Pattern> includePatternsFromString = new ArrayList<>();
    for(String regex : includes.split(","))
    {
      includePatternsFromString.add(Pattern.compile(regex.strip()));
    }
    return includePatternsFromString;
  }

  @Bean
  List<Pattern> excludePatterns() {
    List<Pattern> excludePatternsFromString = new ArrayList<>();
    for(String regex : excludes.split(","))
    {
      excludePatternsFromString.add(Pattern.compile(regex.strip()));
    }
    return excludePatternsFromString;
  }
}
