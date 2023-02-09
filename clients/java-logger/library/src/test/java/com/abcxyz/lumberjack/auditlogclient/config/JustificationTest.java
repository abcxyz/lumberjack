/*
 * Copyright 2022 Lumberjack authors (see AUTHORS file)
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

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class JustificationTest {
  @Spy Justification justification;

  @Test
  public void getPublicKeysEndpoint() {
    String publicKeysEndpoint = "*.example.com";
    justification.setPublicKeysEndpoint(publicKeysEndpoint);
    assertSame(publicKeysEndpoint, justification.getPublicKeysEndpoint());
  }

  @Test
  public void isJustificationEnabled() {
    justification.setEnabled(true);
    assertTrue(justification.isEnabled());

    justification.setEnabled(false);
    assertFalse(justification.isEnabled());
  }

  @Test
  public void allowBreakglass() {
    justification.setAllowBreakglass(true);
    assertTrue(justification.allowBreakglass());

    justification.setAllowBreakglass(false);
    assertFalse(justification.allowBreakglass());
  }

  @Test
  public void validate() {
    String publicKeysEndpoint = "*.example.com";
    justification.setPublicKeysEndpoint(publicKeysEndpoint);
    justification.setEnabled(true);
    justification.validate();

    justification.setEnabled(false);
    justification.validate();
  }

  @Test
  public void validate_Exception() {
    justification.setEnabled(true);
    assertThrows(IllegalArgumentException.class, () -> justification.validate());

    justification.setPublicKeysEndpoint("");
    assertThrows(IllegalArgumentException.class, () -> justification.validate());
  }
}
