# Lumberjack Spec

## Audit logging server (gRPC) API

-   The gRPC service definition is
    [audit_log_agent.proto](protos/v1alpha1/audit_log_agent.proto)
-   The log request definition is
    [audit_log_request.proto](protos/v1alpha1/audit_log_request.proto)
    -   Notice that the essential log payload has type
        `google.cloud.audit.AuditLog` which is the same as Google
        [Cloud Audit Logs](https://github.com/googleapis/googleapis/blob/master/google/cloud/audit/audit_log.proto).

## Audit logging client config

For the canonical config spec, reference comments in
[config.go](clients/go/apis/v1alpha1/config.go). The config file must be written
in `yaml` format.

### Explicit audit logging examples

Config the client to send audit logs

-   whose request principal is not a IAM service account
-   to an audit logging server at `audit-logging.example.com:443`

```yaml
version: v1alpha1
backend:
  address: audit-logging.example.com:443
condition:
  regex:
    exclude: ".iam.gserviceaccount.com$"
```

Config the client to audit log all requests and write the logs to the same audit
logging server.

```yaml
version: v1alpha1
backend:
  address: audit-logging.example.com:443
# No condition means audit log all requests.
```

### Auto audit logging in gRPC examples

Blocks `security_context` and `rules` are only used in auto audit logging setup.

Config the client to

-   look up authentication info (request principal) from a JWT without
    validation (assuming application code already handles it)
-   and enable audit logging for all gRPC method

```yaml
version: v1alpha1
backend:
  address: audit-logging.example.com:443
condition:
  regex:
    exclude: ".iam.gserviceaccount.com$"
security_context:
  from_raw_jwt:
  - key: authorization
    prefix: "Bearer "
rules:
- selector: *
```

Config the client to

-   look up authentication info from *multiple* JWT locations
-   and enable audit logging *only* for gRPC method `/com.example.Foo/Bar`

```yaml
version: v1alpha1
backend:
  address: audit-logging.example.com:443
condition:
  regex:
    exclude: ".iam.gserviceaccount.com$"
security_context:
  from_raw_jwt:
  - key: authorization
    prefix: "Bearer "
  - key: x-jwt-assertion # Also look up JWT in metadata with key x-jwt-assertion
rules:
- selector: /com.example.Foo/Bar
  Directive: AUDIT_REQUEST_AND_RESPONSE # Audit both req and resp
```

Config the client to also verify the JWT (warning: not implemented).

```yaml
version: v1alpha1
backend:
  address: audit-logging.example.com:443
condition:
  regex:
    exclude: ".iam.gserviceaccount.com$"
security_context:
  from_raw_jwt:
  - key: authorization
    prefix: "Bearer "
    jwks:
      endpoint: https://example.com/jwks
rules:
- selector: /com.example.Foo/Bar
  Directive: AUDIT_REQUEST_AND_RESPONSE # Audit both req and resp
```

### Overwritable configs

The following configs can be overwritten with env vars.

-   `AUDIT_CLIENT_BACKEND_ADDRESS` - to overwrite backend address
-   `AUDIT_CLIENT_BACKEND_INSECURE_ENABLED` - to force using insecure connection
-   `AUDIT_CLIENT_BACKEND_IMPERSONATE_ACCOUNT` - to impersonate a service
