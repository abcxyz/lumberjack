**Lumberjack is not an official Google product.**

## Audit Logging Shell (Java)

### About

This project contains the basic service that can be triggered to create Audit logs by using a Java Audit Client service. When no Audit Logging Client implementation is provided, it defaults to `DemoAuditLoggingClient` contained in this project.

The main goal of this "shell" app is to provide the wrapper app that imports the Java Audit Client in automated tests and during development if manual testing is needed.

### How to Build & Run the Logging Shell app

#### Prerequisites:

- [Apache Maven](https://maven.apache.org/install.html)
- [Docker](https://docs.docker.com/get-docker/)
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- [Enable Artifact Registry API and create a Docker repo](https://cloud.google.com/artifact-registry/docs/docker/quickstart)


#### Steps:

1.  Clean the workspace and build the package from the project directory (where `pom.xml` is located):

    ```sh
    mvn clean package
    ```

1.  Package into a container and push the container to Artifact Registry:

    ```sh
    docker buildx build \
      --file "server_app.dockerfile" \
      --tag "${REPO_REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/logging-shell" \
      --push \
      .
    ```

1.  Deploy to Cloud Run:

    ```sh
    gcloud run deploy ${CLOUD_RUN_SERVICE_NAME} \
    --image=${REPO_REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/logging-shell \
    --memory=2048Mi \
    --region=us-west1 \
    --project=${PROJECT_ID} \
    --quiet
    ```

1.  Create a log by triggering the deployed service:

    ```sh
    curl -H "Authorization: Bearer $(gcloud auth print-identity-token )" \
    "${SERVICE_URL_RETURNED_IN_PREVIOUS_STEP}?trace_id=${ID_STRING}"
    ```

1.  Audit log should appear in Logs Explorer in Cloud Console with `lumberjack_trace_id` as `${ID_STRING}` from previous step.

#### Notes:
- `PROJECT_ID` is the "Project ID" of the Google Cloud Platform Project the app will be deployed to.
- `REPO_REGION` is the Artifact registry repo region.
- `REPO_NAME` is the Artifact registry repo name.
- `CLOUD_RUN_SERVICE_NAME` is the name of the Cloud Run service name, e.g. `logging-shell`.
- `SERVICE_URL_RETURNED_IN_PREVIOUS_STEP` is obtained as a result of the `gcloud run deploy` step.
- `ID_STRING` can be any string to be used as trace id.
