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
    "pubsub.googleapis.com",
  ])

  project = var.project_id

  service            = each.value
  disable_on_destroy = false

  depends_on = [
    google_project_service.resourcemanager,
  ]
}

resource "google_pubsub_topic" "audit_logs_topic" {
  project = var.project_id

  name = var.topic_id
}

resource "google_pubsub_subscription" "audit_logs_sub" {
  project = var.project_id

  name  = var.subscription_id
  topic = google_pubsub_topic.audit_logs_topic.name

  // Maximize retention duration 7d.
  message_retention_duration = "604800s"
  retain_acked_messages      = true
  ack_deadline_seconds       = 30

  expiration_policy {
    // Don't specify ttl so the subscription never expires.
    // https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/pubsub_subscription
    ttl = ""
  }

  retry_policy {
    minimum_backoff = "1s"
  }

  // TODO: add deadletter settings.
}
