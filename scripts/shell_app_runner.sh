#!/bin/bash
# Copyright 2021 Lumberjack authors (see AUTHORS file)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -eEuo pipefail

ROOT="$(cd "$(dirname "$0")/.." &>/dev/null; pwd -P)"

TF_CI_DIR=${ROOT}/terraform/ci-run
SERVICE_NAME=shell-app-${RANDOM}
BUILD_COMMAND=$1

SHELL_APP_PROJECT_ID=ci-e2e-app-0
BACKEND_PROJECT_ID=ci-e2e-server
BIGQUERY_DATASET_QUERY=audit_logs.audit_abcxyz_data_access
AUDIT_CLIENT_BACKEND_ADDRESS=audit-logging-bi4j6tgkkq-uc.a.run.app:443

terraform -chdir=${TF_CI_DIR} init
terraform -chdir=${TF_CI_DIR} apply -auto-approve \
  -var="project_id=${SHELL_APP_PROJECT_ID}" \
  -var="service_name=${SERVICE_NAME}" \
  -var='env_vars={"AUDIT_CLIENT_BACKEND_ADDRESS":"'${AUDIT_CLIENT_BACKEND_ADDRESS}'"}' \
  -var='build_commands={"app":"'${BUILD_COMMAND}'"}' \
  -var="use_random_tag=true"

clean_up() {
  terraform -chdir=${TF_CI_DIR} destroy -auto-approve \
    -var="project_id=${SHELL_APP_PROJECT_ID}" \
    -var="service_name=${SERVICE_NAME}" \
    -var='build_commands={"app":"'${BUILD_COMMAND}'"}'
}

trap clean_up EXIT

export HTTP_ENDPOINTS=$(terraform -chdir=${TF_CI_DIR} output -json instance_addresses)

cd ${ROOT}/integration
go test github.com/abcxyz/lumberjack/integration/testrunner \
  -id-token=$(gcloud auth print-identity-token) \
  -project-id=${BACKEND_PROJECT_ID} \
  -dataset-query=${BIGQUERY_DATASET_QUERY}
