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

# Run the java formatter against all content of the repo
cd "$(dirname "$0")/.."

VERSION="1.13.0"
JAR_NAME="google-java-format-${VERSION}-all-deps.jar"

if [ ! -f .local/$JAR_NAME ]
then
  mkdir .local
  curl -LJ "https://github.com/google/google-java-format/releases/download/v${VERSION}/${JAR_NAME}" -o .local/$JAR_NAME
  chmod a+x .local/$JAR_NAME
fi

java -jar .local/$JAR_NAME -i **/*
