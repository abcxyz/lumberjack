package com.abcxyz.lumberjack.auditlogclient;

import com.google.cloud.audit.AuditLog;

/** Intended to hold static utils to ease use of our clients. */
public final class AuditLogs {
  private AuditLogs() {}

  public static AuditLog.Builder getBuilderFromContext() {
    return AuditLoggingServerInterceptor.AUDIT_LOG_CTX_KEY.get();
  }
}
