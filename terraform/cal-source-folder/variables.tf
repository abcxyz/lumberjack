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

variable "log_sink_name" {
  type        = string
  default     = "audit-logs"
  description = "The aggregated log sink name that exports all the cloud audit logs for the folder"
}

variable "destination_project_id" {
  type        = string
  description = "The destination GCP project ID that collects all the audit logs."
}

variable "folder_id" {
  type        = string
  description = "The folder ID for which the aggregated folder level log sink should be created."
}

variable "destination_log_sinks" {
  type = list(object({
    kind = string
    name = string
  }))
  description = "The list of log sink destinations by kind and name. E.g. kind=bigquery, name=[dataset name]."

  validation {
    # At the moment, we only support bigquery sink.
    condition     = !contains([for dest in var.destination_log_sinks : dest.kind == "bigquery" && dest.name != ""], false)
    error_message = "Log sink destination must have kind='bigquery' and a non-empty name."
  }
}
