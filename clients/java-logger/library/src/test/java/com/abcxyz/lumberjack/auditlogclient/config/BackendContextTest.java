package com.abcxyz.lumberjack.auditlogclient.config;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class BackendContextTest {
  @Test
  public void remoteIsEnabled() {
    BackendContext backendContext = new BackendContext();
    backendContext.setAddress("example.com");
    assertThat(backendContext.remoteEnabled()).isTrue();
  }

  @Test
  public void remoteIsDisabled() {
    BackendContext backendContext = new BackendContext();
    assertThat(backendContext.remoteEnabled()).isFalse();
    backendContext.setAddress("");
    assertThat(backendContext.remoteEnabled()).isFalse();
  }
}
