package com.abcxyz.lumberjack.auditlogclient.config;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

@Data
public class CloudLoggingConfiguration {
  @JsonProperty("default_project")
  private boolean defaultProject;
  private String project;
}
