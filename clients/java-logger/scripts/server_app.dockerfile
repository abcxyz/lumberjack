FROM maven:3.8.3-openjdk-11 AS builder
COPY . /src
RUN mvn clean package --no-transfer-progress -f /src/clients/java-logger/pom.xml

FROM gcr.io/distroless/java-debian11:11
COPY integration/testrunner/public_key.json public_key.json
COPY --from=builder /src/clients/java-logger/grpc-test-app/target/grpc-test-app-0.0.1.jar server.jar
CMD ["server.jar"]
