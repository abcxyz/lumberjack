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

variable "folder_parent" {
  type        = string
  description = "The parent to hold the environment. E.g. organizations/102291006291 or folders/300968597098"
}

variable "top_folder_id" {
  type        = string
  description = "The top folder name to hold all the e2e resources."
}

variable "billing_account" {
  type        = string
  description = "The billing account to be linked to projects."
}

variable "apps_count" {
  type        = number
  default     = 1
  description = "The number of app projects to create."
}

variable "enable_all_cal" {
  type        = bool
  default     = false
  description = "Whether to enable all audit log types. If enabled, it could incur a lot of CAL."
}

variable "projects_editors" {
  type        = list(string)
  default     = []
  description = "List of IAM entities that can edit the projects in the env. This should be of the form user:[email], serviceAccount:[email], or group:[email]."
}

variable "audit_log_writers" {
  type        = list(string)
  default     = []
  description = "List of IAM entities that can invoke the audit logging server. This should be of the form user:[email], serviceAccount:[email], or group:[email]."
}

variable "server_image" {
  type        = string
  description = "The server container image."
}

variable "renew_random_tag" {
  type        = bool
  default     = false
  description = "Whether to renew a random tag. If set a new random tag will be assigned and trigger a new build."
}

variable "registry_location" {
  type        = string
  default     = "us"
  description = "The container registry location."
}

variable "service_name" {
  type        = string
  default     = "audit-logging"
  description = "The name for the audit logging server service."
}

variable "enable_pubsub_sink" {
  type        = bool
  default     = false
  description = "Whether to enable a PubSub audit log sink."
}

variable "cal_query_overwrite" {
  type        = string
  default     = ""
  description = "The log query to for the sink."
}
