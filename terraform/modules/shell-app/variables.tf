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
  description = "The GCP project to host the shell app."
}

variable "env_vars" {
  type    = map(string)
  default = {}

  description = "Shell app service environment variables."
}

variable "tag" {
  type        = string
  default     = "init"
  description = "The server container image tag. Changing the tag will trigger a new build."
}

variable "use_random_tag" {
  type        = bool
  default     = false
  description = "If true, use a random tag otherwise use the provided tag via the `tag` variable."
}

variable "artifact_registry_location" {
  type        = string
  default     = "us"
  description = "The artifact registry location."
}

variable "service_name" {
  type        = string
  description = "Name of the service, e.g. go-shell-app or java-shell-app."
}

variable "build_command" {
  type        = string
  description = "Command to call the shell app build script via the relative path to this terraform module, e.g. ../../../clients/java-logger/scripts/build_shell.sh"
}