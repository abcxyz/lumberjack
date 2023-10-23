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

locals {
  # ingestion_backed_client_env_vars = {
  #   "AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS" : "${trimprefix(google_cloud_run_service.server.status[0].url, "https://")}:443",
  #   "AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE" : ".*",
  # }
  cloudlogging_backend_client_env_vars = {
    "AUDIT_CLIENT_BACKEND_CLOUDLOGGING_DEFAULT_PROJECT" : "true",
    "AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE" : ".*",
  }
  short_sha = substr(var.commit_sha, 0, 7)
}

# resource "google_cloud_run_service" "server" {
#   project = var.server_project_id

#   name     = "${var.server_service_name}-${local.short_sha}"
#   location = var.region

#   template {
#     spec {

#       service_account_name = var.server_run_sa

#       containers {
#         image = var.server_image
#       }
#     }
#   }
# }

# resource "google_cloud_run_service_iam_member" "audit_log_writer" {
#   project = var.server_project_id

#   location = google_cloud_run_service.server.location
#   service  = google_cloud_run_service.server.name
#   role     = "roles/run.invoker"
#   member   = "serviceAccount:${var.client_run_sa}"
# }

# resource "google_cloud_run_service" "ingestion_backend_client_services" {
#   for_each = var.client_images

#   project = var.client_project_id

#   name     = "${each.key}-${local.short_sha}-ingestion"
#   location = var.region

#   template {
#     spec {

#       service_account_name = var.client_run_sa

#       containers {
#         image = each.value

#         dynamic "env" {
#           for_each = local.ingestion_backed_client_env_vars

#           content {
#             name  = env.key
#             value = env.value
#           }
#         }
#       }
#     }
#   }
# }

resource "google_cloud_run_service" "cloudlogging_backend_client_services" {
  for_each = var.client_images

  project = var.client_project_id

  name     = "${each.key}-${local.short_sha}-cloudlogging"
  location = var.region

  template {
    spec {

      service_account_name = var.client_run_sa

      containers {
        image = each.value

        dynamic "env" {
          for_each = local.cloudlogging_backend_client_env_vars

          content {
            name  = env.key
            value = env.value
          }
        }
      }
    }
  }
}
