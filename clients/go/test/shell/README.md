## Audit Logging Shell (Go)

The shell app located in `test/shell` is a basic HTTP server that emits an
application audit log using the Lumberjack Go client.

The goal of this shell app is to provide a wrapper that uses our client for
automated and manual testing.

### Prerequisites:

-   [Docker](https://docs.docker.com/get-docker/)
-   [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
-   [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli)
    (Note: Don't use the google3 version)
-   [jq](https://stedolan.github.io/jq/)
-   Make sure you have the following permissions in the `tycho.joonix.net` org:
    -   Folder Editor
    -   Project Creator

### Manual Steps

1.  Authenticate to GCP:

    ```sh
    gcloud auth login --update-adc
    ```

1.  Find one of the existing test environments and change directory to that.
    E.g. `$ROOT/terraform/envs/dev`.

    First, sanity check if you can run Terraform without any problem.

    ```sh
    terraform plan
    ```

    This should return without any error.

    ```sh
    # The audit logging server URL.
    export SERVER_URL=${$(terraform output -raw audit_log_server_url)#"https://"}:443
    # The application project.
    export APP_PROJECT=$(terraform output -json app_projects | jq -r '.[0]')
    # The audit logging server project.
    export SERVER_PROJECT=$(terraform output -raw server_project)
    ```

1.  Execute the following steps from the Lumberjack Go client directory, where
    `go.mod` is located. Build and push the Shell app into a container:

    ```sh
    docker buildx build \
      --file "test/shell/Dockerfile" \
      --tag "us-docker.pkg.dev/${APP_PROJECT}/images/logging-shell:${LDAP}" \
      --push \
      .
    ```

1.  Deploy the Shell app to Cloud Run

    ```sh
    gcloud run deploy ${LDAP}-logging-shell \
    --image=us-docker.pkg.dev/${APP_PROJECT}/images/logging-shell:${LDAP} \
    --memory=512Mi \
    --region=us-west1 \
    --project=${APP_PROJECT} \
    --set-env-vars="AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_INCLUDE=.iam.gserviceaccount.com$,AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS=${SERVER_URL}"
    ```

1.  Create a log with a trace ID by triggering the deployed service:

    ```sh
    export SHELL_APP_URL=$(gcloud run services describe ${LDAP}-logging-shell --platform managed --region us-west1 --format 'value(status.url)')
    curl -H "Authorization: Bearer $(gcloud auth print-identity-token)" ${SHELL_APP_URL}/?trace_id=${ID_STRING}
    ```

1.  View the audit log in BigQuery in project `${SERVER_PROJECT}`.

### Notes:

-   `LDAP` is your ldap used to avoid conflicts with other teammates.
-   `ID_STRING` can be any string to be used as trace id.
