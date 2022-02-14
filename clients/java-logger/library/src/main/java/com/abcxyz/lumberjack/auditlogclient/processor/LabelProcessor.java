package com.abcxyz.lumberjack.auditlogclient.processor;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogMutator;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.inject.Inject;
import java.util.HashMap;
import java.util.Map;
import lombok.AllArgsConstructor;
import lombok.RequiredArgsConstructor;

@AllArgsConstructor(onConstructor = @__({@Inject}))
public class LabelProcessor implements LogMutator {
  private AuditLoggingConfiguration config;

  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) throws LogProcessingException {
    if (config.getLabels() != null) {
      Map<String, String> labels = new HashMap<>();
      // Make a copy to get around unmodifiable map in auditLogRequest
      labels.putAll(auditLogRequest.getLabelsMap());
      for (Map.Entry<String, String> entry : config.getLabels().entrySet()) {
        // Do not overwrite explicitly added labels.
          labels.putIfAbsent(entry.getKey(), entry.getValue());
      }
      return auditLogRequest.toBuilder().putAllLabels(labels).build();
    }
    return auditLogRequest;
  }
}
