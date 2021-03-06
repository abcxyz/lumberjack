/**
 * Copyright 2021 Lumberjack authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

locals {
  default_server_revision_annotations = {
    "autoscaling.knative.dev/maxScale" : "10",
    "run.googleapis.com/sandbox" : "gvisor"
  }
  default_server_service_annotations = {
    "run.googleapis.com/ingress" : "all"
    "run.googleapis.com/launch-stage" : "BETA"
  }
  default_server_env_vars = {
    # At the moment we don't have any default env vars.
  }
}

resource "google_project_service" "resourcemanager" {
  project            = var.project_id
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "services" {
  project = var.project_id
  for_each = toset([
    "containerregistry.googleapis.com",
    "monitoring.googleapis.com",
    "run.googleapis.com",
    "stackdriver.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false

  depends_on = [
    google_project_service.resourcemanager,
  ]
}

resource "google_service_account" "server" {
  count        = var.disable_dedicated_sa ? 0 : 1
  project      = var.project_id
  account_id   = "${var.service_name}-sa"
  display_name = "Audit Logging Server Service Account"
}

resource "google_project_iam_member" "server_roles" {
  for_each = var.disable_dedicated_sa ? toset([]) : toset([
    "roles/cloudtrace.agent",
    "roles/logging.logWriter",
    "roles/monitoring.metricWriter",
    "roles/stackdriver.resourceMetadata.writer",
  ])

  project = var.project_id
  role    = each.key
  member  = "serviceAccount:${google_service_account.server[0].email}"
}

resource "google_cloud_run_service_iam_member" "audit_log_writer" {
  for_each = toset(var.audit_log_writers)

  project  = google_cloud_run_service.server.project
  location = google_cloud_run_service.server.location
  service  = google_cloud_run_service.server.name
  role     = "roles/run.invoker"
  member   = each.key
}

resource "google_cloud_run_service" "server" {
  name     = var.service_name
  location = var.region
  project  = var.project_id

  metadata {
    annotations = merge(
      local.default_server_service_annotations,
      var.server_service_annotations_overrides,
    )
  }

  template {
    spec {
      service_account_name = var.disable_dedicated_sa ? null : google_service_account.server[0].email

      containers {
        image = var.server_image

        resources {
          limits = {
            cpu    = "1000m"
            memory = "1G"
          }
        }

        dynamic "env" {
          for_each = merge(
            local.default_server_env_vars,
            var.server_env_vars,
          )

          content {
            name  = env.key
            value = env.value
          }
        }
      }
    }

    metadata {
      annotations = merge(
        local.default_server_revision_annotations,
        var.server_revision_annotations_overrides,
      )
    }
  }

  autogenerate_revision_name = true

  depends_on = [
    google_project_service.services["run.googleapis.com"],
  ]

  lifecycle {
    ignore_changes = [
      metadata[0].annotations["client.knative.dev/user-image"],
      metadata[0].annotations["run.googleapis.com/client-name"],
      metadata[0].annotations["run.googleapis.com/client-version"],
      metadata[0].annotations["run.googleapis.com/ingress-status"],
      metadata[0].annotations["serving.knative.dev/creator"],
      metadata[0].annotations["serving.knative.dev/lastModifier"],
      metadata[0].labels["cloud.googleapis.com/location"],
      template[0].metadata[0].annotations["client.knative.dev/user-image"],
      template[0].metadata[0].annotations["run.googleapis.com/client-name"],
      template[0].metadata[0].annotations["run.googleapis.com/client-version"],
      template[0].metadata[0].annotations["serving.knative.dev/creator"],
      template[0].metadata[0].annotations["serving.knative.dev/lastModifier"],
    ]
  }
}
