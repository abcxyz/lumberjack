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

#!/bin/bash

set -eu

# Performs Maps-specific post-processing on a .tar.gz archive produced by rule
# java_gapic_assembly_gradle_pkg

# Add gradle publish plugin
#
# Arguments:
#   postprocess_dir: The directory that contains the Java files to postprocess.
add_gradle_publish() {
  postprocess_dir="${1}"
  cat >> "${postprocess_dir}/build.gradle" <<EOF

apply from: "./publish.gradle"
EOF
  echo "INFO: Added gradle publish plugin."
}

# Change group name from cloud
#
# Arguments:
#   postprocess_dir: The directory that contains the Java files to postprocess.
change_group() {
  postprocess_dir="${1}"
  for f in $(find "${postprocess_dir}" -name "*.gradle" -type f); do
    sed -e "s/= 'com\.google\.cloud'/= 'com\.google\.maps'/g" "${f}" > "${f}.new" && mv "${f}.new" "${f}"
    sed -e "s/= 'com\.google\.api\.grpc'/= 'com\.google\.maps'/g" "${f}" > "${f}.new" && mv "${f}.new" "${f}"
done
}

# Main entry point
#
# Arguments:
#   postprocess_dir: The directory that contains the Java files to postprocess.
main() {
  postprocess_dir="$1"

  if [ "${postprocess_dir}" = "" ]; then
    echo "postprocess_dir is required"
    exit 1
  fi

  add_gradle_publish "${postprocess_dir}"
  change_group "${postprocess_dir}"
}

main "$@"
