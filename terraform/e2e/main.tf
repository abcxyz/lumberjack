# Copyright 2023 The Authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

locals {
  application_audit_logs_filter_file = var.application_audit_logs_filter_file == "" ? "${path.module}/../static/application_audit_logs_filter.txt" : var.application_audit_logs_filter_file
  cloud_audit_logs_filter_file       = var.cloud_audit_logs_filter_file == "" ? "${path.module}/../static/cloud_audit_logs_filter.txt" : var.cloud_audit_logs_filter_file
}

resource "google_project_service" "services" {
  for_each = toset([
    "bigquery.googleapis.com",
    "logging.googleapis.com",
    "monitoring.googleapis.com",
    "run.googleapis.com",
    "serviceusage.googleapis.com",
    "stackdriver.googleapis.com",
  ])

  project = var.project_id

  service            = each.value
  disable_on_destroy = false
}

resource "google_bigquery_dataset" "log_storage" {
  project = var.project_id

  dataset_id = var.dataset_id
  location   = var.region

  depends_on = [
    google_project_service.services["bigquery.googleapis.com"],
  ]
}

resource "google_logging_organization_sink" "org_cal_sink" {
  count = var.log_sink_org_id == "" ? 0 : 1

  org_id = var.log_sink_org_id

  name        = "lj-cloud-audit"
  description = "Lumberjack log sink for cloud audit logs."

  destination      = "bigquery.googleapis.com/${google_bigquery_dataset.log_storage.id}"
  filter           = file(local.cloud_audit_logs_filter_file)
  include_children = true

  bigquery_options {
    use_partitioned_tables = true
  }
}

resource "google_logging_organization_sink" "org_aal_sink" {
  count = var.log_sink_org_id == "" ? 0 : 1

  org_id = var.log_sink_org_id

  name        = "lj-app-audit"
  description = "Lumberjack log sink for application audit logs."

  destination      = "bigquery.googleapis.com/${google_bigquery_dataset.log_storage.id}"
  filter           = file(local.application_audit_logs_filter_file)
  include_children = true

  bigquery_options {
    use_partitioned_tables = true
  }
}

resource "google_bigquery_dataset_iam_member" "org_cal_writer" {
  count = var.log_sink_org_id == "" ? 0 : 1

  project = var.project_id

  dataset_id = google_bigquery_dataset.log_storage.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_organization_sink.org_cal_sink[0].writer_identity
}

resource "google_bigquery_dataset_iam_member" "org_aal_writer" {
  count = var.log_sink_org_id == "" ? 0 : 1

  project = var.project_id

  dataset_id = google_bigquery_dataset.log_storage.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_organization_sink.org_aal_sink[0].writer_identity
}

resource "google_logging_folder_sink" "folder_cal_sink" {
  for_each = toset(var.log_sink_folder_ids)

  folder = each.key

  name        = "lj-cloud-audit"
  description = "Lumberjack log sink for cloud audit logs."


  destination      = "bigquery.googleapis.com/${google_bigquery_dataset.log_storage.id}"
  filter           = file(local.cloud_audit_logs_filter_file)
  include_children = true

  bigquery_options {
    use_partitioned_tables = true
  }
}

resource "google_logging_folder_sink" "folder_aal_sink" {
  for_each = toset(var.log_sink_folder_ids)

  folder = each.key

  name        = "lj-app-audit"
  description = "Lumberjack log sink for application audit logs."


  destination      = "bigquery.googleapis.com/${google_bigquery_dataset.log_storage.id}"
  filter           = file(local.application_audit_logs_filter_file)
  include_children = true

  bigquery_options {
    use_partitioned_tables = true
  }
}

resource "google_bigquery_dataset_iam_member" "folder_cal_writer" {
  for_each = toset(var.log_sink_folder_ids)

  project = var.project_id

  dataset_id = google_bigquery_dataset.log_storage.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_folder_sink.folder_cal_sink[each.key].writer_identity
}

resource "google_bigquery_dataset_iam_member" "folder_aal_writer" {
  for_each = toset(var.log_sink_folder_ids)

  project = var.project_id

  dataset_id = google_bigquery_dataset.log_storage.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_folder_sink.folder_aal_sink[each.key].writer_identity
}

resource "google_logging_project_sink" "project_cal_sink" {
  for_each = toset(var.log_sink_project_ids)

  project = each.key

  name        = "lj-cloud-audit"
  description = "Lumberjack log sink for cloud audit logs."


  destination            = "bigquery.googleapis.com/${google_bigquery_dataset.log_storage.id}"
  filter                 = file(local.cloud_audit_logs_filter_file)
  unique_writer_identity = true

  bigquery_options {
    use_partitioned_tables = true
  }
}

resource "google_logging_project_sink" "project_aal_sink" {
  for_each = toset(var.log_sink_project_ids)

  project = each.key

  name        = "lj-app-audit"
  description = "Lumberjack log sink for application audit logs."


  destination            = "bigquery.googleapis.com/${google_bigquery_dataset.log_storage.id}"
  filter                 = file(local.application_audit_logs_filter_file)
  unique_writer_identity = true

  bigquery_options {
    use_partitioned_tables = true
  }
}

resource "google_bigquery_dataset_iam_member" "project_cal_writer" {
  for_each = toset(var.log_sink_project_ids)

  project = var.project_id

  dataset_id = google_bigquery_dataset.log_storage.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_project_sink.project_cal_sink[each.key].writer_identity
}

resource "google_bigquery_dataset_iam_member" "project_aal_writer" {
  for_each = toset(var.log_sink_project_ids)

  project = var.project_id

  dataset_id = google_bigquery_dataset.log_storage.dataset_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_project_sink.project_aal_sink[each.key].writer_identity
}
