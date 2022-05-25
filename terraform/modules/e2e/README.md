**Lumberjack is not an official Google product.**

# E2E Module

## Setup

This module is meant to bootstrap an e2e audit logging solution with a minimal
test setup. The resources will be provisioned are:

*   A top folder under the given org
    *   An `apps` folder under the top folder
    *   1-N app project(s) under the `apps` folder
    *   `ADMIN_READ`, `DATA_READ` and `DATA_WRITE` CAL are enabled for the
        `apps` folder
    *   Folder level log sinks to collect CAL for the `apps` folder
    *   A server project under the top folder
    *   An artifact registry to store the container images
    *   A BigQuery dataset as the audit log storage
    *   A audit logging server deployed as Cloud Run service

The first time the module is applied, an audit logging server image (tagged as
`init`) will be built and published to the artifact registry repo in the server
project. The subsequent runs will *not* publish new images *unless*:

*   A new tag is provided as a var, e.g. `-var='tag="v1"'` in which `v1` is the
    new tag
*   Set auto renew tag as a var, e.g. `-var="auto_renew_tag=true"`. With it, the
    `tag` will be ignored and a new random tag will be generated.

See an usage example in `$REPO_ROOT/terraform/envs/dev`.

## Collecting Audit Logs

After bootstraping the solution:

*   Any cloud audit logs from projects under the `apps` folder will be collected
    in the BigQuery dataset
*   Any application audit logs sent to the audit logging server will be
    collected in the BigQuery dataset
