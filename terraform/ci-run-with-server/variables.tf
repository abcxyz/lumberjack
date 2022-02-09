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
  description = "Name of the service."
}

variable "env_vars" {
  type    = map(string)
  default = {}

  description = "Shell app environment variables."
}

variable "build_commands" {
  type        = map(string)
  description = "List of name/command pairs to call the shell app build script via the relative path to this terraform module, e.g. ../../clients/go/test/shell/build.sh"
}

variable "grpc_build_commands" {
  type        = map(string)
  description = "List of name/command pairs to call the test gRPC app build script via the relative path to this terraform module, e.g. ../../clients/go/test/shell/build.sh"
}

variable "server_project_id" {
  type        = string
  description = "Project ID for the Cloud project where the audit logging backend service is deployed."
}

variable "app_project_id" {
  type        = string
  description = "Project ID for the Cloud project where the audit logging shell app is deployed."
}
