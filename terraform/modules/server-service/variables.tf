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

variable "region" {
  type        = string
  default     = "us-central1"
  description = "The default region for resources in the project; individual resources could have more specific variables defined to specify their region/location"
}

variable "project_id" {
  type        = string
  description = "The GCP project to host the audit logging service."
}

variable "service_name" {
  type        = string
  default     = "audit-logging"
  description = "The name for the audit logging server service."
}

variable "server_image" {
  type        = string
  description = "The audit logging server image."
}

variable "audit_log_writers" {
  type        = list(string)
  default     = []
  description = "List of IAM entities that can invoke the audit log server. This should be of the form user:[email], serviceAccount:[email], or group:[email]."
}

variable "disable_dedicated_sa" {
  type        = bool
  default     = false
  description = "Whether to create a dedicated service account to run the audit logging server."
}

variable "server_service_annotations_overrides" {
  type    = map(string)
  default = {}

  description = "Annotations that applies to all services. Can be used to override default_server_service_annotations."
}

variable "server_revision_annotations_overrides" {
  type    = map(string)
  default = {}

  description = "Annotations that applies to all services. Can be used to override default_server_revision_annotations."
}

variable "server_env_vars" {
  type    = map(string)
  default = {}

  description = "Server service environment overrides."
}
