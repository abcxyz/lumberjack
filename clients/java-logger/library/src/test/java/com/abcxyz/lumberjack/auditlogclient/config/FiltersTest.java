/*
 * Copyright 2022 Lumberjack authors (see AUTHORS file)
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

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.doReturn;
import static org.mockito.Mockito.eq;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class FiltersTest {
  @Spy Filters filters;

  @Test
  public void getsIncludes() {
    String includeString = "*.example.com";
    filters.setIncludes(includeString);
    assertThat(filters.getIncludes()).isEqualTo(includeString);
  }

  @Test
  public void getsIncludes_Env() {
    String includeString = "*.example.com";
    filters.setIncludes(includeString);

    String otherIncludeString = "*.other.com";
    doReturn(otherIncludeString)
        .when(filters)
        .getEnvOrDefault(eq(Filters.PRINCIPAL_INCLUDE_ENV_KEY), any());
    assertThat(filters.getIncludes()).isEqualTo(otherIncludeString);
  }

  @Test
  public void getsExcludes() {
    String excludeString = "*.example.com";
    filters.setExcludes(excludeString);
    assertThat(filters.getExcludes()).isEqualTo(excludeString);
  }

  @Test
  public void getsExcludes_Env() {
    String excludeString = "*.example.com";
    filters.setExcludes(excludeString);

    String otherExcludeString = "*.other.com";
    doReturn(otherExcludeString)
        .when(filters)
        .getEnvOrDefault(eq(Filters.PRINCIPAL_EXCLUDE_ENV_KEY), any());
    assertThat(filters.getExcludes()).isEqualTo(otherExcludeString);
  }
}
