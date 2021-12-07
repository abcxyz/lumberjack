#!/usr/bin/env bash
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

if [ -z "${REPO:-}" ]; then
  echo "✋ Missing REPO!" >&2
  exit 1
fi

if [ -z "${APP_NAME:-}" ]; then
  echo "✋ Missing APP_NAME!" >&2
  exit 1
fi

if [ -z "${TAG:-}" ]; then
  echo "✋ Missing TAG!" >&2
  exit 1
fi

GO_ROOT="$(cd "$(dirname "$0")/../.." &>/dev/null; pwd -P)"
IMAGE_NAME=${REPO}/${APP_NAME}:${TAG}

docker build -f ${GO_ROOT}/test/shell/Dockerfile -t ${IMAGE_NAME} ${GO_ROOT}
docker push ${IMAGE_NAME}
