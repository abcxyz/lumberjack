package com.abcxyz.lumberjack.auditlogclient.config;

import com.abcxyz.lumberjack.auditlogclient.exceptions.AuthorizationException;
import io.grpc.Metadata;
import java.util.Optional;

public interface SecuritySpecification {
  boolean isApplicable(Metadata headers);

  Optional<String> getPrincipal(Metadata headers) throws AuthorizationException;
}
