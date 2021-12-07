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

# This script compiles `audit_log_request.proto` into language-specific code.
#
# To compile the proto, run this script from the Lumberjack root dir.
# For context and `protoc` prerequisites, see the link below:
# https://developers.google.com/protocol-buffers/docs/gotutorial#compiling-your-protocol-buffers
#
# Note: For Java, we use the Maven plugin, https://github.com/os72/protoc-jar-maven-plugin, which
# generates the Java code from protos, thus, explicit code generation via protoc is not needed via
# this script.

# Compile Go code.
protoc -I./third_party/googleapis -I. \
  --go_out=. --go-grpc_out=. \
  --go_opt=module=github.com/abcxyz/lumberjack \
  --go-grpc_opt=module=github.com/abcxyz/lumberjack \
  protos/v1alpha1/audit_log_request.proto protos/v1alpha1/audit_log_agent.proto
