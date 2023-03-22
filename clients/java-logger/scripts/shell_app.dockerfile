FROM maven:3.8.3-openjdk-11 AS builder
COPY . /src
RUN mvn clean package --no-transfer-progress -f /src/clients/java-logger/pom.xml

FROM gcr.io/distroless/java-debian11:11
COPY integration/testrunner/test_public_key.key test_public_key.key
COPY --from=builder /src/clients/java-logger/shell/target/*.jar app.jar
CMD ["app.jar"]
