# Integration Test Runner

**Lumberjack is not an official Google product.**

## Prerequisites

-   Set up an environment with the [e2e module](../../terraform/modules/e2e/)
-   Set up the test apps with the
    [CI run module](../../terraform/modules/ci-run-with-server/)

### About

For each endpoint (either HTTP or gRPC) given to the test, the test runner will
send a request to it to mimic a data access operation, wait and check if the
correct audit log has appeared in the BigQuery (the audit log storage).

## Environment Variables and flags:

-   `HTTP_ENDPOINTS` is the environment variable that contains a JSON list of
    HTTP endpoints to be tested.
-   `GRPC_ENDPOINTS` is the environment variable that contains a JSON list of
    gRPC endpoints to be tested. The gRPC server must implement the
    [test service](../protos/talker.proto).
-   `BIGQUERY_PROJECT_ID` is the flag that contains the Cloud project ID that
    contains the BigQuery dataset.
-   `BIGQUERY_DATASET_QUERY` is the flag that contains the ID of the BigQuery
    dataset to be queried.

See more configs [here](./utils/config.go).

## Run

From the `integration` directory, run `go test
github.com/abcxyz/lumberjack/integration/testrunner -id-token=$(gcloud auth
print-identity-token) -project-id=${PROJECT_ID} -dataset-query=${DATASET_QUERY}`
.

If a service account key exists that is pointed to via the environment variable,
`GOOGLE_APPLICATION_CREDENTIALS`, then the flag, `-id-token`, can be omitted to
execute the runner via the service account.
