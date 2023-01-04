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

# protoc compiles the protobufs.
protoc:
	@protoc -I./third_party/googleapis -I./protos/v1alpha1 \
		--go_out=. --go-grpc_out=. \
		--go_opt=module=github.com/abcxyz/lumberjack \
		--go-grpc_opt=module=github.com/abcxyz/lumberjack \
		audit_log_request.proto audit_log_agent.proto
.PHONY: protoc

# update_third_party updates all the dependencies in third_party.
update_third_party: update_third_party_googleapis
.PHONY: update_third_party

# update_third_party_googleapis fetches the latest upstream version of the
# protos from googleapis.
update_third_party_googleapis:
	@rm -rf third_party/googleapis
	@mkdir -p third_party/googleapis
	@curl -sfLo third_party/googleapis.tgz https://github.com/googleapis/googleapis/archive/refs/heads/master.tar.gz
	@tar -xkf third_party/googleapis.tgz -C third_party/googleapis --strip-components 1
	@rm -rf third_party/googleapis.tgz
.PHONY: update_third_party_googleapis
