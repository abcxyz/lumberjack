# Setup

**Lumberjack is not an official Google product.**

## Log Sinks

GCP [log sinks](https://cloud.google.com/logging/docs/routing/overview) are used
to route Lumberjack audit logs to log storage, either a
[log bucket](https://cloud.google.com/logging/docs/routing/overview#buckets) or
other types of
[destinations](https://cloud.google.com/logging/docs/routing/overview#destinations).

See sample log sinks in Terraform for application-level audit logs in
`google_logging_project_sink` resources
[here](../terraform/modules/server-sink/main.tf).

In case you want all your _cloud_ audit logs to end up in the same log storage,
find similar log sinks in `google_logging_project_sink` resources
[here](../terraform/modules/cal-source-project/main.tf).

## (Optional) Ingestion Service

Build the server with [Dockerfile](../scripts/server.dockerfile). Where you run
the ingestion server doesn't matter as long as it's accessible by the
applications. See [sample deployment](../terraform/modules/server-service/) with
Cloud Run in Terraform.

## E2E

Create an e2e environment using the
[e2e Terraform module](../terraform/modules/e2e/).

```
module "e2e" {
  source        = "github.com/abcxyz/lumberjack/terraform/modules/e2e"
  folder_parent = "folders/YOUR_FOLDER_ID"
  top_folder_id = "my-lumberjack-e2e"

  billing_account = "YOUR_BILLING_ACCOUNT"
}
```