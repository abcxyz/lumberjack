# Clients

**Lumberjack is not an official Google product.**

Lumberjack client consists of the following processors and they run in order:

1.  Validator(s): To validate the audit log requests
2.  Mutator(s): To mutate the audit log requests, e.g. applying common labels
3.  Backend(s): Specifying where to send the logs.

You can use a [config file](./config.md) to ease the client initialization, or
"assemble" a client by yourself in code if you are onboarding to Go client. See
exampels below.

## Go

### Create a client from a config file

```go
opts, err := auditopt.FromConfigFile("path/to/config.yaml")
if err != nil {
  // Handle err
}
client, err := audit.NewClient(opts...)
if err != nil {
  // Handle err
}

// Audit log with the client later on.
// client.Log(ctx, req)
```

### "Assemble" a client in code

**Approach 1**: Create a config in code.

```go
ctx := context.Background()
cfg := &api.Config{
  Backend: &api.Backend{
    CloudLogging: &api.CloudLogging{
      DefaultProject: true,
    },
  },
  Condition: &api.Condition{
    Regex: &api.RegexCondition{
      PrincipalInclude: `@example1\.com$|@example2\.com$`,
    },
  },
  Labels: map[string]string{
    "common_label_1": "foobar",
  }
}

opts, err := auditopt.FromConfig(ctx, cfg)
if err != nil {
  // Handle err
}

client, err := audit.NewClient(opts...)
if err != nil {
  // Handle err
}

// Audit log with the client later on.
// client.Log(ctx, req)
```

**Approach 2**: Assemble processors.

```go
ctx := context.Background()
m, err := filtering.NewPrincipalEmailMatcher(filtering.WithIncludes(`@example1\.com$|@example2\.com$`))
if err != nil {
  // Handle err
}
clp, err := cloudlogging.NewProcessor(ctx)
if err != nil {
  // Handle err
}
lp := &audit.LabelProcessor{DefaultLabels: map[string]string{
  "common_label_1": "foobar",
}}

// If justification is required, create a JVS client.
jvsClient, err := client.NewJVSClient(ctx, &client.JVSConfig{JVSEndpoint: "example.com"})
if err != nil {
  // Handle err
}
jp := justification.NewProcessor(jvsClient)

client, err := audit.NewClient(audit.WithValidator(m), audit.WithBackend(clp), audit.WithMutator(lp), audit.WithMutator(jp))
if err != nil {
  // Handle err
}

// Audit log with the client later on.
// client.Log(ctx, req)
```

This is equivalent to creating a client from the following config file:

```yaml
condition:
  regex:
    principal_include: "@example1\\.com$|@example2\\.com$"
backend:
  cloudlogging:
    default_project: true
labels:
  common_label_1: foobar
justification:
  enabled: true
  public_keys_endpoint: "example.com"
  allow_breakglass: false
```

### Extend

Say if you want to add custom "mutator". Implement the `audit.LogProcessor`
interface and provide it as a mutator as an `audit.Option`.

```go
type MyMutator struct {}

func (m *MyMutator) Process(ctx context.Context, req *api.AuditLogRequest) error {
  // Do your own mutation
}

opts, err := auditopt.FromConfigFile("path/to/config.yaml")
if err != nil {
  // Handle err
}

// Create your own mutator
mm = &MyMutator{}

client, err := audit.NewClient(audit.WithMutator(mm), opts...)
if err != nil {
  // Handle err
}

// Audit log with the client later on.
// client.Log(ctx, req)
```

## Java

### Create a client from a config file

By default, the Java client
[loads the config as a resource](https://docs.oracle.com/en/java/javase/11/docs/api/java.base/java/lang/ClassLoader.html#getResource\(java.lang.String\))
named as `audit_logging.yml`. E.g. `src/main/resources/audit_logging.yml` in
your source code. You can override it with env var `AUDIT_CLIENT_CONFIG_NAME`.
For example, if `AUDIT_CLIENT_CONFIG_NAME` is set to be `my_config.yml`, it is
pointing to `src/main/resources/my_config.yml`.

The example below is using [Guice](https://github.com/google/guice).

```java
Injector injector = Guice.createInjector(new AuditLoggingModule());

// For gRPC services, use AuditLoggingServerInterceptor.
AuditLoggingServerInterceptor interceptor =
        injector.getInstance(AuditLoggingServerInterceptor.class);

// Otherwise, use LoggingClient directly.
LoggingClient client = injector.getInstance(LoggingClient.class);
```

If you are using [Sping](https://spring.io/), you can use Guice to create the
logging client and inject it as Bean.

```java
@Bean
LoggingClient loggingClient() {
  Injector injector = Guice.createInjector(new AuditLoggingModule());
  return injector.getInstance(LoggingClient.class);
}
```

### "Assemble" a client in code

Currently we don't support "assembling" a client in Java code. Please file an
issue if you require such capability.

### Extend

Due to all the log processors are initialized via injection, we currently don't
support vending custom log processors to the audit log client. Please file an
issue if you need such capability.
