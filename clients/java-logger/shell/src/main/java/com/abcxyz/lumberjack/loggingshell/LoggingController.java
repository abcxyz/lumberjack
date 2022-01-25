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

import com.abcxyz.lumberjack.auditlogclient.LoggingClient;
import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessingException;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogType;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.cloud.audit.AuditLog;
import com.google.cloud.audit.AuthenticationInfo;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.Struct;
import com.google.protobuf.Value;
import com.google.protobuf.util.JsonFormat;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Random;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestAttribute;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

/** Endpoints for the shell app that imports/uses the Audit Logging client library. */
@Slf4j
@RequiredArgsConstructor
@RestController
public class LoggingController {
  static final String TRACE_ID_PARAMETER_KEY = "trace_id";
  private static final String SERVICE_NAME = "java-shell-app";

  private static final String METHOD_NAME = "abcxyz.LoggingShellService.v1.GetPeopleById";
  private static final String METADATA_FIELD_NAME = "data";
  private static final Map<String, String> METADATA_MAP =
      Map.of(
          "fieldA", "valueA",
          "fieldB", "valueB",
          "fieldC", "valueC");

  private static final List<String> LOCATIONS = List.of("home", "office", "ski resort", "beach");
  private static final List<String> PETS =
      List.of("shiba inu", "husky", "tuxedo Cat", "siberian cat");
  private static final List<String> HOBBIES = List.of("biking", "racing", "running", "skiing");

  private final Map<String, String> peopleInfo = new HashMap<>();

  private final Random random = new Random();
  private final ObjectMapper objectMapper;)
  private final LoggingClient loggingClient;

  @GetMapping("/people/{id}")
  @ResponseStatus(value = HttpStatus.OK)
  String getPeopleById(
      @PathVariable String id,
      @RequestParam(value = TRACE_ID_PARAMETER_KEY) Optional<String> traceId,
      @RequestAttribute(TokenInterceptor.INTERCEPTOR_USER_EMAIL_KEY) String userEmail)
      throws LogProcessingException, JsonProcessingException, InvalidProtocolBufferException {

    // This is the resource that is accessed via this endpoint.
    String person = getPerson(id);

    // Convert the response to a protobuf struct in order add it to the audit log.
    Struct.Builder responseStructBuilder = Struct.newBuilder();
    JsonFormat.parser().merge(person, responseStructBuilder);

    // Create the audit log record.
    AuditLogRequest record =
        AuditLogRequest.newBuilder()
            .setPayload(
                AuditLog.newBuilder()
                    .setServiceName(SERVICE_NAME)
                    .setMethodName(METHOD_NAME)
                    .setResourceName("/people/" + id)
                    .setMetadata(
                        Struct.newBuilder()
                            .putFields(
                                METADATA_FIELD_NAME,
                                Value.newBuilder()
                                    .setStringValue(objectMapper.writeValueAsString(METADATA_MAP))
                                    .build())
                            .build())
                    .setAuthenticationInfo(
                        AuthenticationInfo.newBuilder().setPrincipalEmail(userEmail).build())
                    .setResponse(responseStructBuilder))
            .setType(LogType.DATA_ACCESS)
            .putLabels(TRACE_ID_PARAMETER_KEY, traceId.orElse(""))
            .build();

    // Log the record.
    loggingClient.log(record);

    log.info("Logged successfully with trace id: {} for user: {}", traceId, userEmail);
    return person;
  }

  private String getPerson(String id) throws JsonProcessingException {
    if (!peopleInfo.containsKey(id)) {
      // Randomly create a record about the given person.
      Map<String, String> personDetailsMap =
          Map.of(
              "id",
              id,
              "location",
              LOCATIONS.get(random.nextInt(LOCATIONS.size())),
              "favorite_pet",
              PETS.get(random.nextInt(LOCATIONS.size())),
              "hobby",
              HOBBIES.get(random.nextInt(LOCATIONS.size())));
      peopleInfo.put(id, objectMapper.writeValueAsString(personDetailsMap));
    }
    return peopleInfo.get(id);
  }
}
