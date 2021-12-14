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
  description = "The destination GCP project ID that stores the audit logs."
}

variable "topic_id" {
  type        = string
  default     = "audit_logs"
  description = "The id used to create the PubSub topic to publish audit logs."
}

variable "subscription_id" {
  type        = string
  default     = "audit_logs_sub"
  description = "The id used to create the PubSub pull subscription to subscribe audit logs."
}

variable "subscribers" {
  type        = list(string)
  default     = []
  description = "List of IAM entities that are allowed to subscribe audit logs."
}

resource "google_project_service" "resourcemanager" {
  project            = var.project_id
  service            = "cloudresourcemanager.googleapis.com"
  disable_on_destroy = false
}

resource "google_project_service" "services" {
  project = var.project_id
  for_each = toset([
    "pubsub.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false

  depends_on = [
    google_project_service.resourcemanager,
  ]
}

output "destination_log_sink" {
  value = {
    kind       = "pubsub"
    project_id = var.project_id
    name       = google_pubsub_topic.audit_logs_topic.name
  }
}
