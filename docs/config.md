# Client Configuration

**Lumberjack is not an official Google product.**

Lumberjack client reads config from a `yaml` file (default path:
`/etc/lumberjack/config.yaml`). Basic config fields can be provided (or
overwritten) by env vars. Alternatively, you can construct a client in code
without a config file (see [examples](./clients.md)).

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

By defualt, the client won't return error when logging is failed to avoid
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

## (WIP) Justification

TODO

## Supported env vars

| ENV VAR name                                      | Description              |
| ------------------------------------------------- | ------------------------ |
| AUDIT_CLIENT_BACKEND_CLOUDLOGGING_DEFAULT_PROJECT | Audit logging directly   |
:                                                   : to cloud logging in the  :
:                                                   : default project          :
| AUDIT_CLIENT_BACKEND_CLOUDLOGGING_PROJECT         | Audit logging directly   |
:                                                   : to cloud logging in the  :
:                                                   : given project            :
| AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS               | Audit logging to an      |
:                                                   : ingestion gRPC service   :
:                                                   : in the given address     :
| AUDIT_CLIENT_BACKEND_REMOTE_INSECURE_ENABLED      | Audit logging to an      |
:                                                   : ingestion gRPC service   :
:                                                   : insecurely               :
| AUDIT_CLIENT_BACKEND_REMOTE_IMPERSONATE_ACCOUNT   | Audit logging to an      |
:                                                   : ingestion gRPC service   :
:                                                   : impersonating the given  :
:                                                   : service account          :
| AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE    | Include the matching     |
:                                                   : request principals in    :
:                                                   : audit logging            :
| AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_EXCLUDE    | Exclude the matching     |
:                                                   : request principals in    :
:                                                   : audit logging            :
| AUDIT_CLIENT_LOG_MODE                             | Whether to fail-close    |
:                                                   : audit logging            :
| AUDIT_CLIENT_JVS_ENDPOINT                         | (Experimental) The JVS   |
:                                                   : JWKs address             :
| AUDIT_CLIENT_REQUIRE_JUSTIFICATION                | (Experimental) Whether   |
:                                                   : to require justification :

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
