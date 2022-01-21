package com.abcxyz.lumberjack.auditlogclient.providers;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.Mockito.spy;

import com.google.api.MonitoredResource;
import com.google.inject.ProvisionException;
import com.google.protobuf.Value;
import java.io.IOException;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class RuntimeInfoValueProviderTests {

  MonitoredResource mr = MonitoredResource.newBuilder()
      .setType("gce_instance")
      .putLabels("project_id", "gcp_project")
      .putLabels("instance_id", "testInstance")
      .putLabels("instance_hostname", "testInstanceName")
      .putLabels("zone", "testZone")
      .build();

  @Test
  void getValueWithNoPlatformMatchReturnsNull() throws IOException {
    RuntimeInfoValueProvider runtimeInfoValueProvider = new RuntimeInfoValueProvider();
    RuntimeInfoValueProvider spyRuntimeInfoValueProvider = spy(runtimeInfoValueProvider);
    Mockito.doReturn(false).when(spyRuntimeInfoValueProvider).isOnGCE();
    Value value = spyRuntimeInfoValueProvider.get();
    assertThat(value).isNull();
  }

  @Test
  void getValueWithGCEPlatformMatchReturns() throws IOException {
    RuntimeInfoValueProvider runtimeInfoValueProvider = new RuntimeInfoValueProvider();
    RuntimeInfoValueProvider spyRuntimeInfoValueProvider = spy(runtimeInfoValueProvider);
    Mockito.doReturn(true).when(spyRuntimeInfoValueProvider).isOnGCE();
    Mockito.doReturn(mr).when(spyRuntimeInfoValueProvider).detectGCEResource();
    Value value = spyRuntimeInfoValueProvider.get();
    assertThat(value).isNotNull();
    assertThat(value.getStructValue().containsFields("type")).isEqualTo(true);
    assertThat(value.getStructValue().containsFields("labels")).isEqualTo(true);
  }

  @Test
  void getValueThrowsException() throws IOException {
    RuntimeInfoValueProvider runtimeInfoValueProvider = new RuntimeInfoValueProvider();
    RuntimeInfoValueProvider spyRuntimeInfoValueProvider = spy(runtimeInfoValueProvider);
    Mockito.doThrow(new IOException("IOException")).when(spyRuntimeInfoValueProvider).isOnGCE();
    assertThrows(ProvisionException.class, () -> spyRuntimeInfoValueProvider.get());
  }
}
