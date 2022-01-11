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
TF_ENVS_CI_DIR=${ROOT}/terraform/envs/ci
TF_CI_WITH_SERVER_DIR=${ROOT}/terraform/ci-run-with-server
SERVICE_NAME=ci-with-server-${RANDOM}
GO_BUILD_COMMAND=${ROOT}/clients/go/test/shell/build.sh
JAVA_BUILD_COMMAND=${ROOT}/clients/java-logger/scripts/build_shell.sh

terraform -chdir=${TF_ENVS_CI_DIR} init
terraform -chdir=${TF_ENVS_CI_DIR} apply -auto-approve

SHELL_APP_PROJECT_ID=$(terraform -chdir=${TF_ENVS_CI_DIR} output -raw app_project)
BACKEND_PROJECT_ID=$(terraform -chdir=${TF_ENVS_CI_DIR} output -raw server_project)
BIGQUERY_DATASET_ID=$(terraform -chdir=${TF_ENVS_CI_DIR} output -raw bigquery_dataset_id)

KOKORO_SERVICE_ACCOUNT=kokoro-sa@lumberjack-kokoro.iam.gserviceaccount.com
GCLOUD_ACCOUNT=$(gcloud config get-value account)
echo "GCloudAccount" $GCLOUD_ACCOUNT
if [[ $GCLOUD_ACCOUNT == $KOKORO_SERVICE_ACCOUNT ]]; then
  # When running in Kokoro, impersonate the Kokoro service account to have its email included in the ID token.
  ID_TOKEN=$(gcloud auth print-identity-token --impersonate-service-account=${KOKORO_SERVICE_ACCOUNT} --include-email)
  # Override the default filters that exclude service accounts during integration tests.
  ENV_VARS='env_vars={"AUDIT_CLIENT_FILTER_REGEX_PRINCIPAL_INCLUDE":".iam.gserviceaccount.com$"}'
else
  ID_TOKEN=$(gcloud auth print-identity-token)
  ENV_VARS='env_vars={}'
fi

terraform -chdir=${TF_CI_WITH_SERVER_DIR} init
terraform -chdir=${TF_CI_WITH_SERVER_DIR} apply -auto-approve \
  -var="server_project_id=${BACKEND_PROJECT_ID}" \
  -var="app_project_id=${SHELL_APP_PROJECT_ID}" \
  -var="service_name=${SERVICE_NAME}" \
  -var='build_commands={"go":"'${GO_BUILD_COMMAND}'", "java":"'${JAVA_BUILD_COMMAND}'"}' \
  -var=${ENV_VARS} \
  -var="use_random_tag=true"

clean_up() {
  terraform -chdir=${TF_CI_WITH_SERVER_DIR} destroy -auto-approve \
    -var="server_project_id=${BACKEND_PROJECT_ID}" \
    -var="app_project_id=${SHELL_APP_PROJECT_ID}" \
    -var="service_name=${SERVICE_NAME}" \
    -var='build_commands={"go":"'${GO_BUILD_COMMAND}'", "java":"'${JAVA_BUILD_COMMAND}'"}'
}

trap clean_up EXIT

export HTTP_ENDPOINTS=$(terraform -chdir=${TF_CI_WITH_SERVER_DIR} output -json instance_addresses)
# TODO(b/203448874): Use updated (finalized) log name.
BIGQUERY_DATASET_QUERY=${BIGQUERY_DATASET_ID}.auditlog_gcloudsolutions_dev_data_access

cd ${ROOT}/integration
go test github.com/abcxyz/lumberjack/integration/httptestrunner \
  -id-token=${ID_TOKEN} \
  -project-id=${BACKEND_PROJECT_ID} \
  -dataset-query=${BIGQUERY_DATASET_QUERY}
