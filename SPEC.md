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
    exclude: ".*\\.iam\\.gserviceaccount\\.com$"
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
Setting these fields *don't* automatically enable auto audit logging. Code
change is still required (TODO: link examples).

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
    exclude: ".*\\.iam\\.gserviceaccount\\.com$"
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
    exclude: ".*\\.iam\\.gserviceaccount\\.com$"
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
    exclude: ".*\\.iam\\.gserviceaccount\\.com$"
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

# Auto Audit Logging Behavior
## Unary Calls
Unary calls are single request/response pairs. In this case, a single audit log will be generated that includes all relevant fields, as well as both the request and response. 

Example (edited for brevity):

```json
{
    ...
    "jsonPayload": {
       ...
      "response": {
        "message": "Hello Some Message",
      },
      "request": {
        "message": "Some Message",
        "target": "e5cd4be1-7143-49c5-9f80-9d6ba779fdc3",
      },
      "resource_name": "e5cd4be1-7143-49c5-9f80-9d6ba779fdc3"
    },
    ...
  }
```

## Streaming Calls

### Server -> Client streaming

A log is created for each response (message sent from server). If a request has
been sent before the request occurred, and we haven't logged that request, we
add it as part of the audit log.

Example (edited for brevity):

```json
  {
    "method_name": "abcxyz.test.Talker/Fibonacci",
    "request": {
      "target": "3548b8c0-6b25-40f0-9a8e-da3ef3d0212d",
      "places": "3.0",
    },
    "response": {
      "position": "1.0",
    },
    "operation": {
      "id": "475c0f1c-a4f3-448c-a147-7ee041a64dda",
      "producer": "abcxyz.test.Talker/Fibonacci",
    }
  },
  {
    "method_name": "abcxyz.test.Talker/Fibonacci",
    "request": {},
    "response": {
      "position": "2.0",
      "value": "1.0",
    },
    "operation": {
      "id": "475c0f1c-a4f3-448c-a147-7ee041a64dda",
      "producer": "abcxyz.test.Talker/Fibonacci",
    }
  },
  {
    "method_name": "abcxyz.test.Talker/Fibonacci",
    "request": {},
    "response": {
      "position": "3.0",
      "value": "1.0",
    },
    "operation": {
      "id": "475c0f1c-a4f3-448c-a147-7ee041a64dda",
      "producer": "abcxyz.test.Talker/Fibonacci",
    }
  }
```

In this example, the client sent a single request (for 3 places of fibonacci)
and the server responded with 3 separate responses. Each of those responses have
been audit logged, and the first response includes the request in addition to
the response.

### Client -> Server streaming

We attempt to pair requests (message from client) with responses in a
best-effort fashion. However, if another request comes in before a response
occurs, we log the previously unlogged requests without responses attached.

Example (edited for brevity):

```json
  {
    "method_name": "abcxyz.test.Talker/Addition",
    "request": {
      "target": "cff5c025-09d9-4892-91a8-3a6ec8ca3060",
      "addend": "1.0"
    },
    "response": null,
    "operation": {
      "id": "9e428c00-6e4f-4dd6-bd75-3c8dd83afe6c",
      "producer": "abcxyz.test.Talker/Addition",
    }
  },
  {
    "method_name": "abcxyz.test.Talker/Addition",
    "request": {
      "target": "cff5c025-09d9-4892-91a8-3a6ec8ca3060",
      "addend": "2.0"
    },
    "response": null,
    "operation": {
      "id": "9e428c00-6e4f-4dd6-bd75-3c8dd83afe6c",
      "producer": "abcxyz.test.Talker/Addition",
    }
  },
 {
    "method_name": "abcxyz.test.Talker/Addition",
    "request": {
      "target": "cff5c025-09d9-4892-91a8-3a6ec8ca3060",
      "addend": "3.0"
    },
    "response": {
      "sum": "6"
    },
    "response": null,
    "operation": {
      "id": "9e428c00-6e4f-4dd6-bd75-3c8dd83afe6c",
      "producer": "abcxyz.test.Talker/Addition",
    }
  }
```

In this example, the client sent 3 numbers for the server to add up (1, 2, 3).
Each of those requests got an audit log, and the last request additionally was
paired with the response (6).

### Bi-Directional Streaming

Bi-directional streaming is handled similarly to both of the above, in that we
try to best-effort pair requests and responses, but we only log each request and
response once.

TODO: add demonstrating example once we have a bi-directional streaming integ
test

### Operation Fields

In the examples above we have included the "operation" fields. These fields are
automatically added, and the id + producer within the operation block form a
unique key that can be used to correlate all request/response values within a
single connection/streaming session.
