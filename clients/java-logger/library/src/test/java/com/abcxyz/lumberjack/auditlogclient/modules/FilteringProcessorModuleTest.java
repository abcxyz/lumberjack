package com.abcxyz.lumberjack.auditlogclient.modules;

import static org.assertj.core.api.Assertions.assertThat;

import com.abcxyz.lumberjack.auditlogclient.processor.FilteringProcessor;
import com.google.inject.Guice;
import com.google.inject.Injector;
import org.junit.jupiter.api.Test;

class FilteringProcessorModuleTest {
  private static Injector injector() {
    return Guice.createInjector(new FilteringProcessorModule());
  }

  @Test
  public void providedFilteringProcessorDoesNotHaveExcludeFilterSetByDefault() {
    FilteringProcessor filteringProcessor = injector().getInstance(FilteringProcessor.class);
    assertThat(filteringProcessor.getExcludePatterns()).isEmpty();
  }

  @Test
  public void providedFilteringProcessorDoesNotHaveIncludeFilterSetByDefault() {
    FilteringProcessor filteringProcessor = injector().getInstance(FilteringProcessor.class);
    assertThat(filteringProcessor.getIncludePatterns()).isEmpty();
  }
}
