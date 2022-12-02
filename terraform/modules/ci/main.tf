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
  ingestion_backed_client_env_vars = {
    "AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS" : "${trimprefix(module.server_service.audit_log_server_url, "https://")}:443",
    "AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE" : ".*",
  }
  cloudlogging_backend_client_env_vars = {
    "AUDIT_CLIENT_BACKEND_CLOUDLOGGING_DEFAULT_PROJECT" : "true",
    "AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE" : ".*",
  }
  short_sha = substr(var.commit_sha, 0, 7)
}

module "server_service" {
  source       = "../server-service"
  project_id   = var.server_project_id
  server_image = var.server_image
  service_name = "${var.server_service_name}-${local.short_sha}"
  region       = var.region

  // Disable dedicated service account for audit logging server.
  // Otherwise a new service account will be created per CI run
  // and cause unnecessary resource waste.
  disable_dedicated_sa = true
}

resource "google_cloud_run_service" "ingestion_backend_client_services" {
  for_each = var.client_images

  name     = "${each.key}-${local.short_sha}-ingestion"
  project  = var.client_project_id
  location = var.region

  template {
    spec {
      containers {
        image = each.value

        resources {
          limits = {
            memory = "1G"
          }
        }

        dynamic "env" {
          for_each = local.ingestion_backed_client_env_vars

          content {
            name  = env.key
            value = env.value
          }
        }
      }
    }
  }
}

resource "google_cloud_run_service" "cloudlogging_backend_client_services" {
  for_each = var.client_images

  name     = "${each.key}-${local.short_sha}-cloudlogging"
  project  = var.client_project_id
  location = var.region

  template {
    spec {
      containers {
        image = each.value

        resources {
          limits = {
            memory = "1G"
          }
        }

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