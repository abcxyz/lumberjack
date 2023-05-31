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

variable "project_id" {
  description = "The GCP project to host the log storage."
  type        = string
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "The default log storage location."
}

variable "dataset_id" {
  type        = string
  default     = "audit_logs"
  description = "The dataset id used to create the BigQuery dataset as the audit log storage."
}

variable "log_sink_project_ids" {
  description = "The GCP projects where to create the audit log sink. Omit to create no project log sink."
  type        = list(string)
  default     = []
}

variable "log_sink_folder_ids" {
  description = "The GCP folders where to create the audit log sink. Omit to create no folder log sink."
  type        = list(string)
  default     = []
}

variable "log_sink_org_id" {
  description = "The GCP org where to create the audit log sink. Omit to create no org log sink."
  type        = string
  default     = ""
}

variable "application_audit_logs_filter_file" {
  description = "File path to application audit logs filter."
  type        = string
  default     = ""
}

variable "cloud_audit_logs_filter_file" {
  description = "File path to cloud audit logs filter."
  type        = string
  default     = ""
}
