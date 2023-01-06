FROM maven:3.8.7-eclipse-temurin-17 AS builder
COPY . /src
RUN mvn clean package --no-transfer-progress -f /src/clients/java-logger/pom.xml

FROM gcr.io/distroless/java17-debian11:11
COPY --from=builder /src/clients/java-logger/grpc-test-app/target/grpc-test-app-0.0.1.jar server.jar
CMD ["server.jar"]