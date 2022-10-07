# Lumberjack for gRPC services

**Lumberjack is not an official Google product.**

The auto audit logging can be added to your gRPC service via gRPC interceptor.

## Config

Config blocks `security_context` and `rules` are only used in auto audit logging
setup. Setting these fields *don't* automatically enable auto audit logging.
Code change is still required (see below).

## Examples

Config the client to

-   look up authentication info (request principal) from a JWT without
    validation (assuming application code already handles it)
-   and enable audit logging for *all* gRPC method

```yaml
version: v1alpha1
backend:
  cloudlogging:
    default_project: true
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
  cloudlogging:
    default_project: true
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

## Go interceptor

```go
// Similar to how you would create an audit client.
interceptor, err := audit.NewInterceptor(auditopt.InterceptorFromConfigFile(ctx, auditopt.DefaultConfigFilePath)
if err != nil {
  // handle err
}
defer func() {
  if err := interceptor.Stop(); err != nil {
    // handle err
  }
}()
// Add the interceptors to the gRPC server.
s := grpc.NewServer(grpc.UnaryInterceptor(interceptor.UnaryInterceptor), grpc.StreamInterceptor(interceptor.StreamInterceptor))
```

In case you have other interceptors of your own, use
[`ChainUnaryInterceptor`](https://pkg.go.dev/google.golang.org/grpc#ChainUnaryInterceptor)
and
[`ChainStreamInterceptor`](https://pkg.go.dev/google.golang.org/grpc#ChainStreamInterceptor)
instead.

## Java interceptor

```java
// This interceptor clubs unary interceptor and stream interceptor into one.
Injector injector = Guice.createInjector(new AuditLoggingModule());
AuditLoggingServerInterceptor interceptor = injector.getInstance(AuditLoggingServerInterceptor.class);

// Add the interceptor to the gRPC server.
server = ServerBuilder.forPort(port).intercept(interceptor).build();
```

## Behaviors

## Unary Calls

Unary calls are single request/response pairs. In this case, a single audit log
will be generated that includes all relevant fields, as well as both the request
and response.

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

TODO([#156](https://github.com/abcxyz/lumberjack/issues/156)): add demonstrating
example once we have a bi-directional streaming integ test

### Operation Fields

In the examples above we have included the "operation" fields. These fields are
automatically added, and the id + producer within the operation block form a
unique key that can be used to correlate all request/response values within a
single connection/streaming session.The id is a randomly generated UUID input by
the interceptor, and the producer is the full method name.
