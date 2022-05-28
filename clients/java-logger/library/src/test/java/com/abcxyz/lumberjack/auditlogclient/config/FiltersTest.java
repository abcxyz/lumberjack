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
