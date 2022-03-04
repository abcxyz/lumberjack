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

resource "google_folder" "top_folder" {
  display_name = var.top_folder_id
  parent       = var.folder_parent
}

resource "google_folder" "apps_folder" {
  display_name = "${var.top_folder_id}-apps"
  parent       = google_folder.top_folder.name
}

resource "google_folder_iam_audit_config" "apps_audit_config" {
  count   = var.enable_all_cal ? 1 : 0
  folder  = google_folder.apps_folder.name
  service = "allServices"
  audit_log_config {
    log_type = "ADMIN_READ"
  }
  audit_log_config {
    log_type = "DATA_READ"
  }
  audit_log_config {
    log_type = "DATA_WRITE"
  }
}

resource "google_project" "server_project" {
  name            = "${var.top_folder_id}-server"
  project_id      = "${var.top_folder_id}-server"
  folder_id       = google_folder.top_folder.name
  billing_account = var.billing_account
}

resource "google_project_iam_member" "server_project_editor" {
  for_each = toset(var.projects_editors)
  project  = google_project.server_project.project_id
  role     = "roles/editor"
  member   = each.value
}

# Give the default compute engine service account in each app project
# the permission to invoke the audit logging sever in the server project.
# Given the project level roles/run.invoker to simplify the e2e env set up and avoid the IAM propagation delay
# which may cause flakiness. Ideally, project level roles/run.invoker is not needed. We only need roles/run.invoker for the audit logging server.
resource "google_project_iam_member" "audit_log_writer_iam" {
  count   = var.apps_count
  project = google_project.server_project.project_id
  role    = "roles/run.invoker"
  member  = "serviceAccount:${google_project.app_project[count.index].number}-compute@developer.gserviceaccount.com"

  depends_on = [
    google_project_service.app_project_services,
  ]
}

resource "google_project" "app_project" {
  count           = var.apps_count
  name            = "${var.top_folder_id}-app-${count.index}"
  project_id      = "${var.top_folder_id}-app-${count.index}"
  folder_id       = google_folder.apps_folder.name
  billing_account = var.billing_account
}

locals {
  app_projects_editors = flatten([
    for i in range(var.apps_count) : [
      for e in var.projects_editors : {
        index  = i
        editor = e
      }
    ]
  ])
}

locals {
  app_projects_services = flatten([
    for i in range(var.apps_count) : [
      for s in toset([
        "run.googleapis.com",
        "compute.googleapis.com",
        "artifactregistry.googleapis.com",
        "iamcredentials.googleapis.com",
        ]) : {
        index   = i
        service = s
      }
    ]
  ])
}

resource "google_project_iam_member" "app_project_editor" {
  for_each = { for e in local.app_projects_editors : "${e.editor}-${e.index}" => e }
  project  = google_project.server_project.project_id
  role     = "roles/editor"
  member   = each.value.editor
}

resource "google_project_service" "app_project_serviceusage" {
  count              = var.apps_count
  project            = google_project.app_project[count.index].project_id
  service            = "serviceusage.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "app_project_resourcemanager" {
  count              = var.apps_count
  project            = google_project.app_project[count.index].project_id
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false

  depends_on = [
    google_project_service.app_project_serviceusage,
  ]
}

resource "google_project_service" "app_project_services" {
  for_each           = { for s in local.app_projects_services : "${s.service}-${s.index}" => s }
  project            = google_project.app_project[each.value.index].project_id
  service            = each.value.service
  disable_on_destroy = false

  depends_on = [
    google_project_service.app_project_resourcemanager
  ]
}

resource "google_artifact_registry_repository" "app_project_image_registry" {
  provider = google-beta

  count         = var.apps_count
  location      = var.registry_location
  project       = google_project.app_project[count.index].project_id
  repository_id = "images"
  description   = "Container Registry for the images."
  format        = "DOCKER"

  depends_on = [
    google_project_service.app_project_services,
  ]
}

// Create the log storage in the same server project.
module "log_storage" {
  source     = "../bigquery-destination"
  project_id = google_project.server_project.project_id

  depends_on = [
    google_project_service.server_project_services,
  ]
}

module "pubsub_sink" {
  count = var.enable_pubsub_sink ? 1 : 0

  source     = "../pubsub-destination"
  project_id = google_project.server_project.project_id

  depends_on = [
    google_project_service.server_project_services,
  ]
}

module "server_sink" {
  source     = "../server-sink"
  project_id = google_project.server_project.project_id
  destination_log_sinks = concat(
    [module.log_storage.destination_log_sink],
    module.pubsub_sink[*].destination_log_sink,
  )

  depends_on = [
    google_project_service.server_project_services,
  ]
}

resource "google_project_service" "server_project_serviceusage" {
  project            = google_project.server_project.project_id
  service            = "serviceusage.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "server_project_resourcemanager" {
  project            = google_project.server_project.project_id
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false

  depends_on = [
    google_project_service.server_project_serviceusage,
  ]
}

module "monitoring_dashboards" {
  source       = "../monitoring"
  project_id   = google_project.server_project.project_id
  service_name = var.service_name
  dataset_id   = module.log_storage.destination_log_sink.name
}

resource "google_project_service" "server_project_services" {
  project = google_project.server_project.project_id
  for_each = toset([
    "serviceusage.googleapis.com",
    "artifactregistry.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false

  depends_on = [
    google_project_service.app_project_resourcemanager,
  ]
}

resource "google_artifact_registry_repository" "image_registry" {
  provider = google-beta

  location      = var.registry_location
  project       = google_project.server_project.project_id
  repository_id = "images"
  description   = "Container Registry for the images."
  format        = "DOCKER"

  depends_on = [
    google_project_service.server_project_services,
  ]
}

resource "null_resource" "build" {
  triggers = {
    "tag" = local.tag
  }

  provisioner "local-exec" {
    environment = {
      PROJECT_ID = google_project.server_project.project_id
      TAG        = local.tag
      REPO       = "${var.registry_location}-docker.pkg.dev/${google_project.server_project.project_id}/images/lumberjack"
    }

    command = "${path.module}/../../../clients/go/scripts/build.sh"
  }

  depends_on = [
    google_artifact_registry_repository.image_registry,
  ]
}

module "server_service" {
  source       = "../server-service"
  project_id   = google_project.server_project.project_id
  server_image = "${var.registry_location}-docker.pkg.dev/${google_project.server_project.project_id}/images/lumberjack/server:${local.tag}"
  service_name = var.service_name

  # Give the list of IAM entities in the input variables
  # the permission to invoke the audit logging server.
  audit_log_writers = var.audit_log_writers

  depends_on = [
    google_project_service.server_project_services,
    null_resource.build,
  ]
}

module "folder_sink" {
  source          = "../cal-source-folder"
  folder_id       = google_folder.apps_folder.name
  query_overwrite = var.cal_query_overwrite
  destination_log_sinks = concat(
    [module.log_storage.destination_log_sink],
    module.pubsub_sink[*].destination_log_sink,
  )
}

output "audit_log_server_url" {
  value = module.server_service.audit_log_server_url
}

output "app_projects" {
  # value = google_project.app_project.project_id
  value = toset([
    for p in google_project.app_project : p.project_id
  ])
}

output "server_project" {
  value = google_project.server_project.project_id
}

output "bigquery_dataset_id" {
  value = module.log_storage.destination_log_sink.name
}
