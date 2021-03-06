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
TF_CI_WITH_SERVER_DIR=${ROOT}/terraform/modules/ci-run-with-server
SERVICE_NAME=ci-with-server-${RANDOM}
GO_BUILD_COMMAND=${ROOT}/clients/go/test/shell/build.sh
GO_GRPC_BUILD_COMMAND=${ROOT}/clients/go/test/grpc-app/build.sh
JAVA_BUILD_COMMAND=${ROOT}/clients/java-logger/scripts/build_shell.sh
JAVA_GRPC_BUILD_COMMAND=${ROOT}/clients/java-logger/scripts/build_server.sh

# Hardcode these values.
# Re-applying the CI env in each CI run might cause unexpected changes being applied to the CI env.
SHELL_APP_PROJECT_ID=github-ci-app-0
BACKEND_PROJECT_ID=github-ci-server
BIGQUERY_DATASET_ID=audit_logs

CI_SERVICE_ACCOUNT=gh-access-sa@lumberjack-dev-infra.iam.gserviceaccount.com
GCLOUD_ACCOUNT=$(gcloud config get-value account)
if [[ $GCLOUD_ACCOUNT == $CI_SERVICE_ACCOUNT ]]; then
  # When running in CI, impersonate the service account to have its email included in the ID token.
  ID_TOKEN=$(gcloud auth print-identity-token --impersonate-service-account=${CI_SERVICE_ACCOUNT} --include-email)
else
  ID_TOKEN=$(gcloud auth print-identity-token)
fi
# Override the default filters to include all principals.
ENV_VARS='env_vars={"AUDIT_CLIENT_CONDITION_REGEX_PRINCIPAL_INCLUDE":".*"}'

terraform -chdir=${TF_CI_WITH_SERVER_DIR} init
terraform -chdir=${TF_CI_WITH_SERVER_DIR} apply -auto-approve \
  -var="server_project_id=${BACKEND_PROJECT_ID}" \
  -var="app_project_id=${SHELL_APP_PROJECT_ID}" \
  -var="service_name=${SERVICE_NAME}" \
  -var='build_commands={"go":"'${GO_BUILD_COMMAND}'", "java":"'${JAVA_BUILD_COMMAND}'"}' \
  -var='grpc_build_commands={"go":"'${GO_GRPC_BUILD_COMMAND}'", "java":"'${JAVA_GRPC_BUILD_COMMAND}'"}' \
  -var=${ENV_VARS} \
  -var="use_random_tag=true"

clean_up() {
  terraform -chdir=${TF_CI_WITH_SERVER_DIR} destroy -auto-approve \
    -var="server_project_id=${BACKEND_PROJECT_ID}" \
    -var="app_project_id=${SHELL_APP_PROJECT_ID}" \
    -var="service_name=${SERVICE_NAME}" \
    -var='build_commands={"go":"'${GO_BUILD_COMMAND}'", "java":"'${JAVA_BUILD_COMMAND}'"}' \
    -var='grpc_build_commands={"go":"'${GO_GRPC_BUILD_COMMAND}'", "java":"'${JAVA_GRPC_BUILD_COMMAND}'"}'
}

trap clean_up EXIT

export HTTP_ENDPOINTS=$(terraform -chdir=${TF_CI_WITH_SERVER_DIR} output -json instance_addresses)
export GRPC_ENDPOINTS=$(terraform -chdir=${TF_CI_WITH_SERVER_DIR} output -json grpc_addresses)
# TODO(b/203448874): Use updated (finalized) log name.
BIGQUERY_DATASET_QUERY=${BIGQUERY_DATASET_ID}.audit_abcxyz_data_access

cd ${ROOT}/integration
go test github.com/abcxyz/lumberjack/integration/testrunner\
  -id-token=${ID_TOKEN} \
  -project-id=${BACKEND_PROJECT_ID} \
  -dataset-query=${BIGQUERY_DATASET_QUERY}
