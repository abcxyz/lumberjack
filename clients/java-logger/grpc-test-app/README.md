**Lumberjack is not an official Google product.**

Protos and server/client are largely based on the examples found in the grpc-java repo. https://github.com/grpc/grpc-java

### How to Build & Run the grpc test app

#### Prerequisites:

- [Apache Maven](https://maven.apache.org/install.html)
- [Docker](https://docs.docker.com/get-docker/)
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
- [Enable Artifact Registry API and create a Docker repo](https://cloud.google.com/artifact-registry/docs/docker/quickstart)

#### Steps:
1.  Build the libraries into jar files (from java-logger/ directory run below command):

    ```
    mvn clean package
    ```
    A jar file is generated in the target folder of each subfolder after this.

2.  Package into a container and push the container to Artifact Registry from
    the root of repository:

    ```sh
    docker buildx build \
      --file "clients/java-logger/scripts/server_app.dockerfile" \
      --tag "${REPO_REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/grpc-test-app" \
      --push \
      .
    ```

1.  Deploy to Cloud Run:

    ```sh
    gcloud run deploy ${CLOUD_RUN_SERVICE_NAME} \
    --image=${REPO_REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}/grpc-test-app \
    --memory=2048Mi \
    --region=us-west1 \
    --project=${PROJECT_ID} \
    --set-env-vars=AUDIT_CLIENT_BACKEND_CLOUDLOGGING_DEFAULT_PROJECT=true \
    --quiet
    ```

1.  Send requests from the client:
    ```
    java -cp clients/java-logger/grpc-test-app/target/grpc-test-app-0.0.1.jar abcxyz.lumberjack.test.talker.TalkerClient ${SERVICE_URL_RETURNED_IN_PREVIOUS_STEP} 443 $(gcloud auth print-identity-token)
    ```

1.  Audit log should appear in Logs Explorer in Cloud Console with service name `abcxyz.test.Talker` from previous step.

#### Notes:
- `PROJECT_ID` is the "Project ID" of the Google Cloud Platform Project the app will be deployed to.
- `REPO_REGION` is the Artifact registry repo region.
- `REPO_NAME` is the Artifact registry repo name.
- `CLOUD_RUN_SERVICE_NAME` is the name of the Cloud Run service name, e.g. `grpc-test-app`.
- `SERVICE_URL_RETURNED_IN_PREVIOUS_STEP` is obtained as a result of the `gcloud run deploy` step.