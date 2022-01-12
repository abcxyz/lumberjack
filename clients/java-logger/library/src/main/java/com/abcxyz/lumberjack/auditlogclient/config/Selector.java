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

import static com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogType;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Collection;
import java.util.Optional;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.NonNull;

/**
 * This class is intended to specify which methods should be audit logged, and specific
 * configuration for those methods. Multiple selectors could match a single method as we allow wild
 * cards in the patterns.
 *
 * <p>Example patterns:
 * "*"
 * "com.example.*"
 * "com.example.Hello"
 */
@Data
@AllArgsConstructor
@NoArgsConstructor
public class Selector {
  private static final String WILD_CARD = "*";

  @JsonProperty("selector")
  @NonNull
  private String pattern;

  private Directive directive;
  private LogType logType;

  public LogType getLogType() {
    return logType == null ? LogType.DATA_ACCESS : logType;
  }

  public Directive getDirective() {
    return directive == null ? Directive.AUDIT : directive;
  }

  public int getLength() {
    return pattern.length();
  }

  /** Determines if this selector should be applied to the method. */
  public boolean isApplicable(String methodIdentifier) {
    if (pattern.equals(WILD_CARD)) {
      return true;
    } else if (pattern.endsWith(WILD_CARD)) {
      return methodIdentifier.startsWith(pattern.substring(0, pattern.length() - 1));
    } else {
      return methodIdentifier.startsWith(pattern);
    }
  }

  /**
   * This method uses the length of the selector to determine which selector is most relevant.
   *
   * <p> e.g. com.example.Hello > com.example.* > *
   */
  public static Optional<Selector> returnMostRelevant(
      String methodIdentifier, Collection<Selector> selectors) {
    int longest = 0;
    Optional<Selector> mostRelevant = Optional.empty();
    for (Selector selector : selectors) {
      if (selector.isApplicable(methodIdentifier) && selector.getLength() > longest) {
        longest = selector.getLength();
        mostRelevant = Optional.of(selector);
        if (longest == methodIdentifier.length()) {
          // immediately return if the method identifier is the same length as selector
          return mostRelevant;
        }
      }
    }
    return mostRelevant;
  }
}
