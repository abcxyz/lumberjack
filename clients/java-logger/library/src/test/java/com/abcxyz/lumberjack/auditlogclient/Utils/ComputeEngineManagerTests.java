package com.abcxyz.lumberjack.auditlogclient.Utils;


import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.auditlogclient.utils.CloudFunctionManager;
import com.abcxyz.lumberjack.auditlogclient.utils.ComputeEngineManager;
import com.abcxyz.lumberjack.auditlogclient.utils.RuntimeInfoUtils;
import com.google.api.MonitoredResource;
import java.io.IOException;
import java.net.URL;
import java.net.URLConnection;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Mockito;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class ComputeEngineManagerTests {

  @Mock
  RuntimeInfoUtils runtimeInfoUtils;

  @InjectMocks
  ComputeEngineManager computeEngineManager;

  @Test
  void detectGCEResourceReturnsResourceOnValidMetadata() {
    Mockito.doReturn("testProject").when(runtimeInfoUtils).getProjectId();
    Mockito.doReturn("testInstanceId").when(runtimeInfoUtils).getInstanceId();
    Mockito.doReturn("testInstanceName").when(runtimeInfoUtils).getInstanceName();
    Mockito.doReturn("testZone").when(runtimeInfoUtils).getZone();
    MonitoredResource mr = computeEngineManager.detectGCEResource();
    assertThat(mr.containsLabels("project_id")).isTrue();
    assertThat(mr.containsLabels("instance_name")).isTrue();
  }

  @Test
  void detectGCEResourceThrowsExceptionResourceOnInvalidMetadata() {
    Mockito.doThrow(new IllegalArgumentException("Exception")).when(runtimeInfoUtils).getProjectId();
    assertThrows(IllegalArgumentException.class, () -> computeEngineManager.detectGCEResource());
  }
}
