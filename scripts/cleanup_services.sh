#!/usr/bin/env bash
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
