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

# Run `terraform init` and `terraform validate`.

module "cal_project_sources" {
  for_each   = toset(["fake-proj1", "fake-proj2", "fake-proj3"])
  source     = "../cal-source-project"
  project_id = each.key
  destination_log_sinks = [
    {
      kind       = "bigquery"
      project_id = "lumberjack-dest"
      name       = "auditlogs-all"
    },
    {
      kind       = "bigquery"
      project_id = "lumberjack-dest"
      name       = "auditlogs-secondary"
    }
  ]
}

module "cal_folder_sources" {
  for_each  = toset(["fake-folder1", "fake-folder2"])
  source    = "../cal-source-folder"
  folder_id = each.key
  destination_log_sinks = [
    {
      kind       = "bigquery"
      project_id = "lumberjack-dest"
      name       = "auditlogs-all"
    },
    {
      kind       = "bigquery"
      project_id = "lumberjack-dest"
      name       = "auditlogs-secondary"
    }
  ]
}

module "server-sink" {
  source     = "../server-sink"
  project_id = "lumberjack-server"
  destination_log_sinks = [
    {
      kind       = "bigquery"
      project_id = "lumberjack-dest"
      name       = "auditlogs-all"
    }
  ]
}

module "server-service" {
  source       = "../server-service"
  project_id   = "lumberjack-server"
  server_image = "gcr.io/lumberjack-server/lumberjack/server:fake"
}

module "bigquery-destination" {
  source     = "../bigquery-destination"
  project_id = "bigquery-destination"
}

module "pubsub-destination" {
  source     = "../pubsub-destination"
  project_id = "pubsub-destination"
}

provider "google" {
  user_project_override = true
}

provider "google-beta" {
  user_project_override = true
}
