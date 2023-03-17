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

resource "google_project_service" "resourcemanager" {
  project = var.project_id

  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "services" {
  for_each = toset([
    "logging.googleapis.com",
  ])

  project = var.project_id

  service            = each.value
  disable_on_destroy = false

  depends_on = [
    google_project_service.resourcemanager,
  ]
}

resource "google_logging_project_sink" "bigquery_sink" {
  for_each = { for dest in var.destination_log_sinks : dest.name => dest if dest.kind == "bigquery" }

  project = var.project_id

  name        = format("%s-%s", var.log_sink_name, each.value.name)
  destination = "bigquery.googleapis.com/projects/${each.value.project_id}/datasets/${each.value.name}"

  filter = var.query_overwrite != "" ? var.query_overwrite : <<-EOT
  LOG_ID("audit.abcxyz/unspecified") OR
  LOG_ID("audit.abcxyz/activity") OR
  LOG_ID("audit.abcxyz/data_access")
  EOT

  unique_writer_identity = true
  bigquery_options {
    use_partitioned_tables = true
  }

  depends_on = [
    google_project_service.services["logging.googleapis.com"],
  ]
}

resource "google_bigquery_dataset_iam_member" "bigquery_sink_memeber" {
  for_each = { for dest in var.destination_log_sinks : dest.name => dest if dest.kind == "bigquery" }

  project = each.value.project_id

  dataset_id = each.value.name
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_project_sink.bigquery_sink[each.value.name].writer_identity
}

resource "google_logging_project_sink" "pubsub_sink" {
  for_each = { for dest in var.destination_log_sinks : dest.name => dest if dest.kind == "pubsub" }

  project = var.project_id

  name        = format("ps-%s-%s", var.log_sink_name, each.value.name)
  destination = "pubsub.googleapis.com/projects/${each.value.project_id}/topics/${each.value.name}"

  filter = var.query_overwrite != "" ? var.query_overwrite : <<-EOT
  LOG_ID("audit.abcxyz/unspecified") OR
  LOG_ID("audit.abcxyz/activity") OR
  LOG_ID("audit.abcxyz/data_access")
  EOT

  unique_writer_identity = true

  depends_on = [
    google_project_service.services["logging.googleapis.com"],
  ]
}

resource "google_pubsub_topic_iam_member" "pubsub_sink_member" {
  for_each = { for dest in var.destination_log_sinks : dest.name => dest if dest.kind == "pubsub" }

  project = each.value.project_id

  topic  = each.value.name
  role   = "roles/pubsub.publisher"
  member = google_logging_project_sink.pubsub_sink[each.value.name].writer_identity
}
