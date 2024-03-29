# Copyright 2022 Google LLC
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

# Use distroless for ca certs.
FROM gcr.io/distroless/static AS distroless


# Use a scratch image to host our binary.
FROM scratch
COPY --from=distroless /etc/passwd /etc/passwd
COPY --from=distroless /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY server /server

# Normally we would set this to run as "nobody".
# But goreleaser builds the binary locally and sometimes it will mess up the permission
# and cause "exec user process caused: permission denied".
#
# USER nobody

# Run the web service on container startup.
ENV PORT 8080
ENTRYPOINT ["/server"]