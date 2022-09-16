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

variable "server_image" {
  type        = string
  description = "Server image name"
}

variable "server_project_id" {
  type        = string
  description = "Project ID for the Cloud project where the audit logging backend service is deployed."
}

variable "client_project_id" {
  type        = string
  description = "Project ID for the Cloud project where the client services are deployed."
}

variable "grpc_client_images" {
  type        = list(string)
  default     = []
  description = "gRPC Client image names of implementations to deploy"
}

variable "http_client_images" {
  type        = list(string)
  default     = []
  description = "HTTP Client image names of implementations to deploy"
}

variable "docker_repo" {
  type    = string
  default = "us-docker.pkg.dev/lumberjack-dev-infra/images"
}

variable "docker_tag" {
  type = string
}
