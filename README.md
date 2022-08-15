# Lumberjack

**Lumberjack is not an official Google product.**

Lumberjack helps your applications on Google Cloud to write audit logs similar
to Cloud Audit Logs and provides building blocks to centrally collect the audit
logs for analysis.

## Use Case

Audit logs are special logs that record *when* and *who* called *which*
application and accessed *what* data. And *why* the access was necessary.

The typical use case of Lumberjack is for an organization's
[insider risk (or insider threat)](https://en.wikipedia.org/wiki/Insider_threat)
program. The organization will require its applications to write audit logs
whenever there are *employees* calling the applications to access *user data*.

## Components

Lumberjack consists of the following components:

-   [Ingestion service](./cmd/server/)
-   [Go Client](./clients/go)
-   [Java Logger](./clients/java-logger)

See manuals for client [configuration](./docs/config.md) and
[usage](./docs/clients.md).

## Setup

Lumberjack supports three setups:

-   **Log to ingestion service**: This setup writes audit logs to the ingestion
    service. It allows you to *centrally* control the audit log ingestion and
    add common log processing logic.
-   **Log to GCP Cloud Logging**: This setup writes audit logs directly to Cloud
    Logging. It eliminates the need to *run and operate* the ingestion service.
-   **Log to stdout** (WIP): This setup writes audit logs to stdout. It's
    suitable for environments where access to ingestion service and Cloud
    Logging are not possible.
