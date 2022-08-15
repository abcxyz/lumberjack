# Clients

**Lumberjack is not an official Google product.**

Lumberjack client consists of the following processors and they run in order:

1.  Validator(s): To validate audit log requests
2.  Mutator(s): To mutate audit log request, e.g. applying common labels
3.  Backend(s): Specifying where to send the logs.

You can use a [config file](./config.md) to ease the client initialization or
"assemble" a client by yourself in code. See exampels below.

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

### "Assemable" a client in code

```go
m, err := filtering.NewPrincipalEmailMatcher(filtering.WithIncludes(`@example1\.com$|@example2\.com$`))
if err != nil {
  // Handle err
}
p, err := cloudlogging.NewProcessor(context.Background())
if err != nil {
  // Handle err
}
l := &audit.LabelProcessor{DefaultLabels: map[string]string{
  "common_label_1": "foobar",
}}

client, err = audit.NewClient(audit.WithValidator(m), audit.WithBackend(p), audit.WithMutator(l))
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
```

### Exend

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

TODO
