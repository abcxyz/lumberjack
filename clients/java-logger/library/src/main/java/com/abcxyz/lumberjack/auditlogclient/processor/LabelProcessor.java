package com.abcxyz.lumberjack.auditlogclient.processor;

import com.abcxyz.lumberjack.auditlogclient.config.AuditLoggingConfiguration;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogMutator;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.google.inject.Inject;
import java.util.HashMap;
import java.util.Map;
import lombok.AllArgsConstructor;

/**
 * Adds labels from the configuration to the audit log request. Does not overwrite existing labels.
 */
@AllArgsConstructor(onConstructor = @__({@Inject}))
public class LabelProcessor implements LogMutator {
  private AuditLoggingConfiguration config;

  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) throws LogProcessingException {
    if (config.getLabels() == null || config.getLabels().isEmpty()) {
      // shortcut if there are no labels to add
      return auditLogRequest;
    }
    AuditLogRequest.Builder builder = auditLogRequest.toBuilder();
    config.getLabels().entrySet().stream()
        .filter(e -> !auditLogRequest.getLabelsMap().containsKey(e.getKey()))
        .forEach(e -> builder.putLabels(e.getKey(), e.getValue()));
    return builder.build();
  }
}
