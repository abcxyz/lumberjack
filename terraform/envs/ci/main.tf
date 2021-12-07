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

terraform {
  backend "gcs" {
    # Bucket is in project "lumberjack-dev-infra"
    # We can reuse this project for CI and sandbox envs with a different prefix.
    bucket = "lumberjack-dev-terraform"
    prefix = "ci"
  }
}

provider "google" {
  user_project_override = true
}

provider "google-beta" {
  user_project_override = true
}

# If we want to release a new image for the audit logging server,
# we can provide a tag, e.g. -var="tag=v1"
variable "tag" {
  type        = string
  default     = "init"
  description = "The server container image tag. Changing the tag will trigger a new build."
}

# When set to true, it will ignore the given tag.
# Instead, it will generate a random UUID as the image tag.
# This is handy and only meant for testing only (e.g. in CI).
variable "renew_random_tag" {
  type        = bool
  default     = false
  description = "Whether to renew a random tag. If set a new random tag will be assigned and trigger a new build."
}


module "e2e" {
  source        = "../../e2e"
  folder_parent = "folders/316290568068"
  top_folder_id = "ci-e2e"

  // The billing account 'Gong Test'.
  billing_account = "016242-61A3FB-F92462"

  tag              = var.tag
  renew_random_tag = var.renew_random_tag
}

# Preemptively give the audit log writer permission to invoke any Cloud Run
# instance in the Server project. While the invoker permission is also granted
# on the individual service level for each service spawned in the server
# project, this way, the IAM propagation delay for individual Cloud Run
# instances during an integration run is aimed to be avoided.
resource "google_project_iam_member" "audit_log_writer" {
  project  = module.e2e.server_project
  role     = "roles/run.invoker"
  member   = "serviceAccount:${module.e2e.audit_log_writer}"
}

output "audit_log_server_url" {
  value = module.e2e.audit_log_server_url
}

output "audit_log_writer" {
  value = module.e2e.audit_log_writer
}

output "server_project" {
  value = module.e2e.server_project
}

output "app_project" {
  value = tolist(module.e2e.app_projects).0
}

output "bigquery_dataset_id" {
  value = module.e2e.bigquery_dataset_id
}
