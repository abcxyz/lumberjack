/*
 * Copyright 2021 Lumberjack authors (see AUTHORS file)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.abcxyz.lumberjack.auditlogclient.config;

/** This enum is used to specify which details are audit logged. */
public enum Directive {
  AUDIT_REQUEST_AND_RESPONSE(true, true),
  AUDIT_REQUEST_ONLY(true, false),
  AUDIT(false, false);

  private final boolean logRequest;
  private final boolean logResponse;

  private Directive(boolean logRequest, boolean logResponse) {
    this.logRequest = logRequest;
    this.logResponse = logResponse;
  }

  public boolean shouldLogRequest() {
    return logRequest;
  }

  public boolean shouldLogResponse() {
    return logResponse;
  }
}
