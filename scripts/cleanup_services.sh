#!/usr/bin/env bash
# Copyright 2022 Lumberjack authors (see AUTHORS file)
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

echo "Starting cleanup."

echo "account details:"
gcloud config get-value account

projectId="github-ci-app-0"
regionId="us-central1"
echo "Setting context $projectId"
gcloud config set project ${projectId}

yesterday=$(date --date="yesterday" +"%Y-%m-%d")
echo "Fetching services to delete older than $yesterday "

echo "Fetching gcloud services."
servicesToDelete=( $(gcloud run services list --filter="metadata.creationTimestamp<${yesterday}" --format="value(SERVICE)" --project="${projectId}") )

for serviceName in "${servicesToDelete[@]}"; do
  gcloud run services delete ${serviceName} --project="${projectId}" --region="${regionId}" --quiet
  echo "Deleted: ${serviceName}"
done

echo "done"
