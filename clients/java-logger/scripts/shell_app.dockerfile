FROM maven:3.8.3-openjdk-17 AS builder
COPY . /src
RUN mvn clean package --no-transfer-progress -f /src/clients/java-logger/pom.xml

FROM gcr.io/distroless/java17-debian11:latest
COPY integration/testrunner/test_jwks test_jwks
COPY --from=builder /src/clients/java-logger/shell/target/*.jar app.jar
CMD ["app.jar"]
