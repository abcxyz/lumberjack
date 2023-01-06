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

package com.abcxyz.lumberjack.loggingshell;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.abcxyz.lumberjack.auditlogclient.LoggingClient;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.mockito.Captor;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.request.MockHttpServletRequestBuilder;

@WebMvcTest(LoggingController.class)
public class LoggingControllerTest {
  private static final String REQUEST_PATH = "/";
  private static final String TEST_TRACE_ID = "testTraceId";
  private static final String TEST_EMAIL = "testEmail";
  private static final MockHttpServletRequestBuilder GET_REQUEST_BUILDER = get(REQUEST_PATH);
  private static final MockHttpServletRequestBuilder GET_REQUEST_BUILDER_WITH_EMAIL =
      get(REQUEST_PATH).requestAttr(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY, TEST_EMAIL);
  private static final MockHttpServletRequestBuilder GET_REQUEST_BUILDER_WITH_EMAIL_AND_TRACE_ID =
      get(REQUEST_PATH)
          .requestAttr(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY, TEST_EMAIL)
          .param(LoggingController.TRACE_ID_PARAMETER_KEY, TEST_TRACE_ID);

  @Autowired private MockMvc mockMvc;
  @MockBean private LoggingClient loggingClient;
  @MockBean private TokenInterceptor interceptor;
  @Captor private ArgumentCaptor<AuditLogRequest> auditLogRequestCaptor;

  @BeforeEach
  void setUp() {
    when(interceptor.preHandle(any(), any(), any())).thenReturn(true);
  }

  @Test
  void loggingWithoutTraceIdCausesBadRequestError() throws Exception {
    mockMvc.perform(GET_REQUEST_BUILDER_WITH_EMAIL).andExpect(status().isBadRequest());
  }

  @Test
  void loggingWithTraceIdReturnsOk() throws Exception {
    mockMvc.perform(GET_REQUEST_BUILDER_WITH_EMAIL_AND_TRACE_ID).andExpect(status().isOk());
  }

  @Test
  void loggingUsesTheEmailInRequestAuthenticationInfo() throws Exception {
    mockMvc.perform(GET_REQUEST_BUILDER_WITH_EMAIL_AND_TRACE_ID);
    verify(loggingClient).log(auditLogRequestCaptor.capture());
    assertThat(
            auditLogRequestCaptor
                .getValue()
                .getPayload()
                .getAuthenticationInfo()
                .getPrincipalEmail())
        .isEqualTo(TEST_EMAIL);
  }

  @Test
  void loggingWithoutInterceptedEmailCausesBadRequestError() throws Exception {
    mockMvc.perform(GET_REQUEST_BUILDER).andExpect(status().isBadRequest());
  }

  @Test
  void loggingWithCustomTraceIdMakesAuditLoggingClientCallWithCustomTraceId() throws Exception {
    mockMvc.perform(GET_REQUEST_BUILDER_WITH_EMAIL_AND_TRACE_ID);
    verify(loggingClient).log(auditLogRequestCaptor.capture());
    assertThat(
            auditLogRequestCaptor
                .getValue()
                .getLabelsMap()
                .getOrDefault(LoggingController.TRACE_ID_PARAMETER_KEY, ""))
        .isEqualTo(TEST_TRACE_ID);
  }
}
