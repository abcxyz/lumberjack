To deploy the server to Cloud Run, run the following command from the Lumberjack Go root:

```
docker build -t us-docker.pkg.dev/noamrabbani-test/images/grpc-talker-server -f test/grpc-app/Dockerfile . && \

docker push us-docker.pkg.dev/noamrabbani-test/images/grpc-talker-server && \

gcloud run deploy grpc-talker-server --image=us-docker.pkg.dev/noamrabbani-test/images/grpc-talker-server --memory=2048Mi --region=us-west1 --project=noamrabbani-test --set-secrets=/secrets/config/auditconfig.yaml=auditconfig:latest
```