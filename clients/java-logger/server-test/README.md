Protos and server/client are largely based on the examples found in the grpc-java repo. https://github.com/grpc/grpc-java

Build:

```
mvn verify
```

Run the server (fill out app project and server url below):
```
export APP_PROJECT=
export REPO=us-docker.pkg.dev/$APP_PROJECT/images
export APP_NAME=java-server-test-$USER
export TAG=init
export SERVER_URL=

./build_server.sh

gcloud run deploy $APP_NAME \
--image=us-docker.pkg.dev/$APP_PROJECT/images/$APP_NAME:init \
--region=us-west1 \
--project=${APP_PROJECT} \
--platform=managed \
--set-env-vars="AUDIT_CLIENT_BACKEND_ADDRESS=${SERVER_URL}"
```

In another terminal, send requests from the client:
```
export SERVICE_URL=
mvn exec:java -Dexec.mainClass=abcxyx.helloworld.HelloWorldClientTls -Dexec.args="${SERVICE_URL} 443 $(gcloud auth print-identity-token)"
```