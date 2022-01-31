## HTTP Endpoints Test Runner

### About

This package's aim is to test the HTTP Endpoints in the Audit logging project.
The main usage of this package is intended to be via automation, but it can be
run locally for dev/testing purposes as well.

### How to Run the Test Runner

#### Prerequisites:

- Audit logging service should be deployed.
- BigQuery database should be in place.
- Log sink that transfers the audit logs to BigQuery database should be working.
- Once these are set, test runner can be executed after providing configuration
  info via environment variables and flags described below.

#### Environment Variables and flags:

- `HTTP_ENDPOINTS` is the environment variable that contains a JSON list of HTTP
  endpoints to be tested.
- `BIGQUERY_PROJECT_ID` is the flag that contains the Cloud project ID that
  contains the BigQuery dataset.
- `BIGQUERY_DATASET_QUERY` is the flag that contains the ID of the BigQuery
  dataset to be queried.

#### Running:

From the `integration` directory,
run `go test github.com/abcxyz/lumberjack/integration/testrunner -id-token=$(gcloud auth print-identity-token) -project-id=${PROJECT_ID} -dataset-query=${DATASET_QUERY}`
.

If a service account key exists that is pointed to via the environment
variable, `GOOGLE_APPLICATION_CREDENTIALS`, then the flag, `-id-token`, can be
omitted to execute the runner via the service account.
