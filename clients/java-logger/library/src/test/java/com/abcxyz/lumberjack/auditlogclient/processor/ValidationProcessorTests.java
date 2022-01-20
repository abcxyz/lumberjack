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

package com.abcxyz.lumberjack.auditlogclient.processor;

import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.junit.jupiter.api.Assertions.assertThrows;

import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogType;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class ValidationProcessorTests {

  @Test
  void auditLogRequestNull() {
    ValidationProcessor validationProcessor = new ValidationProcessor();
    assertThrows(IllegalArgumentException.class, () -> validationProcessor.process(null));
  }

  @Test
  void validatesWhenRequestIsValid() {
    ValidationProcessor validationProcessor = new ValidationProcessor();
    AuditLogRequest record =
        AuditLogRequest.newBuilder()
            .setPayload(
                AuditLog.newBuilder()
                    .setAuthenticationInfo(
                        AuthenticationInfo.newBuilder().setPrincipalEmail("foo").build()))
            .setType(LogType.DATA_ACCESS)
            .build();
    assertDoesNotThrow(() -> validationProcessor.process(record));
  }
}
