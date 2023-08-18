# Client Configuration

**Lumberjack is not an official Google product.**

Default config path

-   Go client: `/etc/lumberjack/config.yaml`
-   Java client: a
    [resource](https://docs.oracle.com/en/java/javase/11/docs/api/java.base/java/lang/ClassLoader.html#getResource\(java.lang.String\))
    with name `audit_logging.yml` (e.g. `src/main/resources/audit_logging.yml`)

Lumberjack clients read config from a yaml file located at the default path or a
path you specify. Basic config fields can be provided (or overwritten) by env
vars. Alternatively, you can construct a client in code without a config file
(see [examples](./clients.md)).

For the canonical config spec, reference comments in
[config.go](clients/go/apis/v1alpha1/config.go).

## Backend

To write audit logs to an ingestion service, add the following block in the
config:

```yaml
backend:
  remote:
    # The address of your ingestion service.
    address: audit-logging.example.com:443
```

To write audit logs to Cloud Logging, add the following block in the config:

```yaml
backend:
  cloudlogging:
    # Use the cloud project where the serice runs
    default_project: true
    # Or to override the project:
    # project: my-logging-project
```

## Condition

Often we only want to audit log human accesses. Condition allows you to
selectively audit log requests.

E.g. To only log requests with principal email ends with `@example.com`, add the
following block in the config:

```yaml
condition:
  regex:
    include: "@example\\.com$"
```

E.g. To *not* log any GCP service account initiated requests, add the following
block in the config:

```yaml
condition:
  regex:
    exclude: ".*\\.iam\\.gserviceaccount\\.com$"
```

## Log Mode

By default, the client won't return error when logging is failed to avoid
breaking the application. However, certain applications might prefer failing the
request if audit logging is failed ("fail-close"). Add the following block in
the config to enable fail-close mode:

```yaml
log_mode: FAIL_CLOSE
```

## Labels

To add static labels to all audit logs, add the following block in the config:

```yaml
labels:
  foo: bar
  abc: xyz
```

## gRPC configs

Please refer to the [gRPC guide](./grpc.md).

## Justification

By default, justifications will not be added to audit logs, even if provided.
When enabling adding justification, `public_keys_endpoint` must be provided to
fetch the public keys which is used to validate JWTs that are passed in through
the `justification_token` header. To enable adding justification information to
audit logs, add the following block in the config:

```yaml
justification:
  enabled: true
  public_keys_endpoint: example.com
  allow_breakglass: false
```

## Supported environment variables

ENV VAR name                                      | Description
------------------------------------------------- | -----------
LUMBERJACK_LOG_LEVEL                              | Verbosity of the lumberjack server logs; valid values are "debug", "warn", "info" (default), "error".
LUMBERJACK_LOG_FORMAT                             | Output format for lumberjack server logs; valid values are "text" or "json" (default).
AUDIT_CLIENT_BACKEND_CLOUDLOGGING_DEFAULT_PROJECT | Audit logging directly to cloud logging in the default project
AUDIT_CLIENT_BACKEND_CLOUDLOGGING_PROJECT         | Audit logging directly to cloud logging in the given project
AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS               | Audit logging to an ingestion gRPC service in the given address
AUDIT_CLIENT_BACKEND_REMOTE_INSECURE_ENABLED      | Audit logging to an ingestion gRPC service insecurely
AUDIT_CLIENT_BACKEND_REMOTE_IMPERSONATE_ACCOUNT   | Audit logging to an ingestion gRPC service impersonating the given service account
AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE    | Include the matching request principals in audit logging
AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_EXCLUDE    | Exclude the matching request principals in audit logging
AUDIT_CLIENT_LOG_MODE                             | Whether to fail-close audit logging
AUDIT_CLIENT_CONFIG_NAME                          | (For Java client only) The config file (e.g. `src/main/resources/${AUDIT_CLIENT_CONFIG_NAME}`) to use
AUDIT_CLIENT_JUSTIFICATION_PUBLIC_KEYS_ENDPOINT   | (Experimental) The JVS JWKs address
AUDIT_CLIENT_JUSTIFICATION_ENABLED                | (Experimental) Whether to enable justification
AUDIT_CLIENT_JUSTIFICATION_ALLOW_BREAKGLASS       | (Experimental) Whether to allow breakglass, ignored if justification is not enabled.

## Examples

```yaml
backend:
  cloudlogging:
    default_project: true
condition:
  regex:
    include: "@example\\.com$"
```

The config above will write audit logs with principal email ends with
`@example.com` to Cloud Logging in the same project. You can achieve the same by
providing the following env vars:

```sh
AUDIT_CLIENT_BACKEND_CLOUDLOGGING_DEFAULT_PROJECT=true
AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE="@example\\.com$"
```
