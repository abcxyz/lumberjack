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

module "shell_app" {
  source = "../shell-app"

  for_each = var.build_commands

  build_command              = each.value
  project_id                 = var.project_id
  service_name               = "${var.service_name}-${each.key}"
  env_vars                   = var.env_vars
  tag                        = var.tag
  use_random_tag             = var.use_random_tag
  region                     = var.region
  artifact_registry_location = var.artifact_registry_location
}
