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
  repo = "${var.artifact_registry_location}-docker.pkg.dev/${var.project_id}/images/lumberjack"
  tag  = var.use_random_tag ? uuid() : var.tag
  default_server_env_vars = {
  }
}

resource "null_resource" "hello_app_build" {
  triggers = {
    "tag" = local.tag
  }

  provisioner "local-exec" {
    environment = {
      REPO     = local.repo
      APP_NAME = var.service_name
      TAG      = local.tag
    }
    command = var.build_command
  }
}

resource "google_cloud_run_service" "hello_app_service" {
  name     = var.service_name
  location = var.region
  project  = var.project_id

  template {
    spec {
      containers {
        image = "${local.repo}/${var.service_name}:${local.tag}"

        resources {
          limits = {
            memory = "1G"
          }
        }

        dynamic "env" {
          for_each = merge(local.default_server_env_vars, var.env_vars)

          content {
            name  = env.key
            value = env.value
          }
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [
    null_resource.hello_app_build,
  ]
}

output "hello_address" {
  value = google_cloud_run_service.hello_app_service.status.0.url
}
