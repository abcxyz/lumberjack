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

output "server_url" {
  value = module.server_service.audit_log_server_url
}

output "grpc_client_urls" {
  value = [
    for _, value in var.grpc_client_images :
    lookup(google_cloud_run_service.grpc_client_services, value).status[0].url
  ]
}

output "http_client_urls" {
  value = [
    for _, value in var.http_client_images :
    lookup(google_cloud_run_service.http_client_services, value).status[0].url
  ]
}
