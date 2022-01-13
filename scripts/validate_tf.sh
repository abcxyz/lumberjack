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

# terraform validation for level 1 folders inside terrform dir
terraform_dirs=$(ls -d terraform/*)

for dir in ${terraform_dirs[@]}; do
  echo "terraform validating $dir"
  terraform -chdir=$dir fmt -check
  terraform -chdir=$dir init -backend=false
  terraform -chdir=$dir validate
done
