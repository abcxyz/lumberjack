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
module "lumberjack" {
  source = "git::https://github.com/abcxyz/lumberjack.git//terraform/e2e?ref=vREPLACE_LUMBERJACK_RELEASE_VERSION"

  project_id = "REPLACE_PROJECT_ID"

  region = "REPLACE_REGION"
  dataset_id = "REPLACE_DATASET_ID"
  log_sink_project_ids = REPLACE_LOG_SINK_PROJECT_IDS
  log_sink_folder_ids = REPLACE_LOG_SINK_FOLDER_IDS
  log_sink_org_id = "REPLACE_SINK_ORG_ID"
  application_audit_logs_filter_file = "REPLACE_APPLICATION_AUDIT_LOGS_FILTER_FILE"
  cloud_audit_logs_filter_file = "REPLACE_AUDIT_LOGS_FILTER_FILE"
}