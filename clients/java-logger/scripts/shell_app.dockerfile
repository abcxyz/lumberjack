FROM maven:3.8.7-amazoncorretto-17 AS builder
COPY . /src
RUN mvn clean package --no-transfer-progress -f /src/clients/java-logger/pom.xml

FROM gcr.io/distroless/java17-debian11
COPY --from=builder /src/clients/java-logger/shell/target/*.jar app.jar
CMD ["app.jar"]
