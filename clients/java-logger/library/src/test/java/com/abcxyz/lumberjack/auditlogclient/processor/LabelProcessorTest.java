package com.abcxyz.lumberjack.auditlogclient.processor;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.doReturn;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.cloud.audit.AuditLog;
import java.util.HashMap;
import java.util.Map;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class LabelProcessorTest {
  @Mock
  AuditLoggingConfiguration auditLoggingConfiguration;
  AuditLog auditLog = AuditLog.newBuilder().build();
  AuditLogRequest auditLogRequest = AuditLogRequest.newBuilder().setPayload(auditLog).build();

  @Test
  public void testProcess() throws LogProcessingException {
    Map<String, String> labels = new HashMap<>();
    labels.put("label1", "value1");
    labels.put("label2", "value2");
    doReturn(labels).when(auditLoggingConfiguration).getLabels();
    LabelProcessor labelProcessor = new LabelProcessor(auditLoggingConfiguration);
    AuditLogRequest output = labelProcessor.process(auditLogRequest);
    assertThat(output.getLabelsMap()).isEqualTo(labels);
  }

  @Test
  public void testProcess_WithExistingLabels() throws LogProcessingException {
    Map<String, String> labels = new HashMap<>();
    labels.put("label1", "value1");
    labels.put("label2", "value2");
    doReturn(labels).when(auditLoggingConfiguration).getLabels();
    LabelProcessor labelProcessor = new LabelProcessor(auditLoggingConfiguration);

    String otherValue = "otherValue1";
    AuditLogRequest request= AuditLogRequest.newBuilder()
        .setPayload(auditLog)
        .putLabels("label1", otherValue)
        .build();
    AuditLogRequest output = labelProcessor.process(request);
    assertThat(output.getLabelsMap()).isNotEqualTo(labels);
    assertThat(output.getLabelsMap().get("label1")).isEqualTo(otherValue);
  }

  @Test
  public void testProcess_WithExistingLabels_Null() throws LogProcessingException {
    doReturn(null).when(auditLoggingConfiguration).getLabels();
    LabelProcessor labelProcessor = new LabelProcessor(auditLoggingConfiguration);

    String otherValue = "otherValue1";
    AuditLogRequest request= AuditLogRequest.newBuilder()
        .setPayload(auditLog)
        .putLabels("label1", otherValue)
        .build();
    AuditLogRequest output = labelProcessor.process(request);
    assertThat(output.getLabelsMap().size()).isEqualTo(1);
    assertThat(output.getLabelsMap().get("label1")).isEqualTo(otherValue);
  }
}
