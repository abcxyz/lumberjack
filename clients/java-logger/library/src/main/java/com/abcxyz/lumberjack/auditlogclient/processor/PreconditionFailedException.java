/*
 * Copyright 2023 Lumberjack authors (see AUTHORS file)
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

package com.abcxyz.lumberjack.auditlogclient.processor;

/** Thrown when log processors wants to stop processing the log request. */
public class PreconditionFailedException extends LogProcessingException {
  public PreconditionFailedException(Throwable cause) {
    super(cause);
  }

  public PreconditionFailedException(String message, Throwable cause) {
    super(message, cause);
  }

  public PreconditionFailedException(String message) {
    super(message);
  }
}
