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


# Fail on any error.
set -eEuo pipefail

# Run all Go tests.
# Go test only works from a Go module.
(cd clients/go && go test -shuffle=on -count=1 -race -timeout=10m -coverprofile=coverage.out ./...)


# terraform validation for level 1 folders inside terrform dir
mapfile -t terraform_dirs < <(ls -d terraform/* || true)

for dir in "${terraform_dirs[@]}"; do
  echo "terraform validating ${dir}"
  terraform -chdir="${dir}" fmt -check
  terraform -chdir="${dir}" init -backend=false
  terraform -chdir="${dir}" validate
done

# Run tests in java/maven projects.
mvn clean test --no-transfer-progress -f clients/java-logger

# Run the integration tests, but not in GitHub Actions yet.
if [[ -v GITHUB_ACTIONS ]]
then
  echo "gh actions not configured for IT"
else
  ./scripts/integration_build.sh
fi
