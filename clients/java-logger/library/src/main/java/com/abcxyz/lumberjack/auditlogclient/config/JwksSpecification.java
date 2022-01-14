package com.abcxyz.lumberjack.auditlogclient.config;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class JwksSpecification {
  private String endpoint;
  private String object;
}
