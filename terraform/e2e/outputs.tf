# Copyright 2023 The Authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

output "dataset_full_id" {
  description = "The BigQuery dataset full id of the log storage."
  value       = google_bigquery_dataset.log_storage.id
}

output "dataset_id" {
  description = "The BigQuery dataset id of the log storage."
  value       = google_bigquery_dataset.log_storage.dataset_id
}
