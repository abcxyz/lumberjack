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
  repo = "${var.artifact_registry_location}-docker.pkg.dev/${var.server_project_id}/images/lumberjack"
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

    command = "${path.module}/../../clients/go/scripts/build.sh"
  }
}

module "server_service" {
  source       = "../server-service"
  project_id   = var.server_project_id
  server_image = "${local.repo}/server:${local.tag}"
  service_name = var.service_name

  depends_on = [
    null_resource.server_build,
  ]
}

locals {
  env_vars = {
    "AUDIT_CLIENT_BACKEND_ADDRESS" : "${trimprefix(module.server_service.audit_log_server_url, "https://")}:443",
  }
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

module "java_grpc_app" {
  source = "../shell-app"

  build_command              = var.java_grpc_build_command
  project_id                 = var.app_project_id
  service_name               = "${var.service_name}-java-grpc"
  env_vars                   = merge(local.env_vars, var.env_vars)
  tag                        = var.tag
  use_random_tag             = var.use_random_tag
  region                     = var.region
  artifact_registry_location = var.artifact_registry_location
}

output "instance_addresses" {
  value = [for key, _ in var.build_commands : module.shell_app[key].instance_address]
}

output "grpc_address" {
  value = module.java_grpc_app.instance_address
}