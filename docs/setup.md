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

In case you want all your *cloud* audit logs to end up in the same log storage,
find similar log sinks in `google_logging_project_sink` resources
[here](../terraform/modules/cal-source-project/main.tf).

## (Optional) Ingestion Service

To build the ingestion service:

```sh
# By default we use Lumberjack CI container registry for the images.
# To override, set the following env var.
# DOCKER_REPO=us-docker.pkg.dev/my-project/images

# goreleaser expects a "clean" repo to release so commit any local changes if
# needed.
git add . && git commit -m "local changes"

# goreleaser expects a tag.
# The tag must be a semantic version https://semver.org/
git tag -f -a v0.0.0-$(git rev-parse --short HEAD)

# Use goreleaser to build the images.
# It should in the end push all the images to the given container registry.
# All the images will be tagged with the git tag given earlier.
goreleaser release -f .goreleaser.docker.yaml --rm-dist
```

Where you run the ingestion server doesn't matter as long as it's accessible by
the applications. See [sample deployment](../terraform/modules/server-service/)
with Cloud Run in Terraform.

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
