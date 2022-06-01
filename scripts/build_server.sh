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

ROOT="$(cd "$(dirname "$0")/.." &>/dev/null; pwd -P)"

if [ -z "${PROJECT_ID:-}" ]; then
  echo "✋ Missing PROJECT_ID!" >&2
  exit 1
fi

if [ -z "${TAG:-}" ]; then
  # TODO(b/203448889): Disallow dirty repo.
  TAG="$(git describe --dirty --always)"
  echo "🎈 Using tag ${TAG}!" >&2
fi

if [ -z "${REPO:-}" ]; then
  # TODO(b/203448996): Use full github repo path.
  REPO="gcr.io/${PROJECT_ID}/lumberjack"
  echo "🎈 Using repo ${REPO}!" >&2
fi

docker build \
  --file="$(dirname "$0")/server.dockerfile" \
  --tag="${REPO}/lumberjack-server:${TAG}" \
  ${ROOT}

docker push "${REPO}/lumberjack-server:${TAG}"
