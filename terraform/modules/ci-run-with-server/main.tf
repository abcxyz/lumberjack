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
  tag  = var.use_random_tag ? uuid() : var.tag
  repo = "${var.artifact_registry_location}-docker.pkg.dev/${var.server_project_id}/images"
  env_vars = {
    "AUDIT_CLIENT_BACKEND_REMOTE_ADDRESS" : "${trimprefix(module.server_service.audit_log_server_url, "https://")}:443",
  }
}

resource "null_resource" "server_build" {
  triggers = {
    "tag" = local.tag
  }

  provisioner "local-exec" {
    environment = {
      PROJECT_ID = var.server_project_id
      TAG        = local.tag
      REPO       = local.repo
    }

    command = "${path.module}/../../../scripts/build_server.sh"
  }
}

module "server_service" {
  source       = "../server-service"
  project_id   = var.server_project_id
  server_image = "${local.repo}/lumberjack-server:${local.tag}"
  service_name = var.service_name

  // Disable dedicated service account for audit logging server.
  // Otherwise a new service account will be created per CI run
  // and cause unnecessary resource waste.
  disable_dedicated_sa = true

  depends_on = [
    null_resource.server_build,
  ]
}

module "shell_app" {
  source = "../shell-app"

  for_each = var.build_commands

  build_command              = each.value
  project_id                 = var.app_project_id
  service_name               = "${var.service_name}-${each.key}"
  env_vars                   = merge(local.env_vars, var.env_vars)
  tag                        = var.tag
  use_random_tag             = var.use_random_tag
  region                     = var.region
  artifact_registry_location = var.artifact_registry_location
}

module "grpc_app" {
  source = "../shell-app"

  for_each = var.grpc_build_commands

  build_command              = each.value
  project_id                 = var.app_project_id
  service_name               = "grpc-${var.service_name}-${each.key}"
  env_vars                   = merge(local.env_vars, var.env_vars)
  tag                        = var.tag
  use_random_tag             = var.use_random_tag
  region                     = var.region
  artifact_registry_location = var.artifact_registry_location
}
