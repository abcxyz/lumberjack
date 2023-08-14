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

# update_third_party updates all the dependencies in third_party.
update_third_party: update_third_party_googleapis
.PHONY: update_third_party

# update_third_party_googleapis fetches the latest upstream version of the
# protos from googleapis.
update_third_party_googleapis:
	@rm -rf third_party/googleapis
	@mkdir -p third_party/googleapis/google/cloud third_party/tmp
	@curl -sfLo third_party/googleapis.tgz https://github.com/googleapis/googleapis/archive/refs/heads/master.tar.gz
	@tar -xkf third_party/googleapis.tgz -C third_party/tmp --strip-components 1
	@cp -rf third_party/tmp/google/cloud/audit third_party/googleapis/google/cloud
	@cp -rf third_party/tmp/google/logging third_party/googleapis/google
	@cp -rf third_party/tmp/google/api third_party/googleapis/google
	@cp -rf third_party/tmp/google/rpc third_party/googleapis/google
	@find third_party/googleapis -type f ! -name '*.proto' -delete
	@rm -rf third_party/tmp
	@rm -rf third_party/googleapis.tgz
.PHONY: update_third_party_googleapis
