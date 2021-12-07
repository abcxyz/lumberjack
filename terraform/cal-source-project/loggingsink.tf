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

resource "google_logging_project_sink" "bigquery_sink" {
  for_each    = { for dest in var.destination_log_sinks : dest.name => dest if dest.kind == "bigquery" }
  name        = format("%s-%s", var.log_sink_name, each.value.name)
  project     = var.project_id
  destination = "bigquery.googleapis.com/projects/${var.destination_project_id}/datasets/${each.value.name}"

  filter = <<-EOT
  LOG_ID("cloudaudit.googleapis.com/activity") OR
  LOG_ID("externalaudit.googleapis.com/activity") OR
  LOG_ID("cloudaudit.googleapis.com/system_event") OR
  LOG_ID("externalaudit.googleapis.com/system_event") OR
  LOG_ID("cloudaudit.googleapis.com/access_transparency") OR
  LOG_ID("externalaudit.googleapis.com/access_transparency") OR
  LOG_ID("cloudaudit.googleapis.com/data_access") OR
  LOG_ID("externalaudit.googleapis.com/data_access")
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
  for_each   = { for dest in var.destination_log_sinks : dest.name => dest if dest.kind == "bigquery" }
  dataset_id = each.value.name
  project    = var.destination_project_id
  role       = "roles/bigquery.dataEditor"
  member     = google_logging_project_sink.bigquery_sink[each.value.name].writer_identity
}
