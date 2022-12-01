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

import com.abcxyz.lumberjack.auditlogclient.processor.LogProcessor.LogBackend;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest.LogType;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequestProto;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.google.cloud.MonitoredResource;
import com.google.cloud.logging.LogEntry;
import com.google.cloud.logging.Logging;
import com.google.cloud.logging.LoggingException;
import com.google.cloud.logging.Payload;
import com.google.inject.Inject;
import com.google.protobuf.InvalidProtocolBufferException;
import com.google.protobuf.util.JsonFormat;
import java.io.UnsupportedEncodingException;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.Collections;
import java.util.Map;
import lombok.AccessLevel;
import lombok.AllArgsConstructor;
import lombok.extern.java.Log;

/** Cloud logging processor to write logs to google cloud */
@Log
@AllArgsConstructor(access = AccessLevel.PRIVATE, onConstructor = @__({@Inject}))
public class CloudLoggingProcessor implements LogBackend {
  private static final String MONITORED_RESOURCE_TYPE = "global";
  private final ObjectMapper mapper;
  private final Logging logging;

  /**
   * Sends the {@link AuditLogRequest} to the google cloud logging
   *
   * @param auditLogRequest Audit log request to be processed/logged using Google Cloud Logging.
   * @return The identical {@link AuditLogRequest} without any updates performed on it.
   * @throws LogProcessingException if {@code auditLogRequest} proto has problems preventing from
   *     being converted to JSON or if the log type is not URL encode-able.
   */
  @Override
  public AuditLogRequest process(AuditLogRequest auditLogRequest) throws LogProcessingException {
    try {
      LogEntry entry =
          LogEntry.newBuilder(
                  Payload.JsonPayload.of(
                      mapper.readValue(
                          JsonFormat.printer().
                                  preservingProtoFieldNames().
                                  print(auditLogRequest.getPayload()),
                          new TypeReference<Map<String, ?>>() {})))
              .setLogName(getLogNameFromLogType(auditLogRequest.getType()))
              .setLabels(auditLogRequest.getLabelsMap())
              .setResource(MonitoredResource.newBuilder(MONITORED_RESOURCE_TYPE).build())
              .build();
      logging.write(Collections.singleton(entry));
      return auditLogRequest;
    } catch (InvalidProtocolBufferException
        | JsonProcessingException
        | UnsupportedEncodingException
        | LoggingException ex) {
      throw new LogProcessingException(ex);
    } finally {
      logging.flush();
    }
  }

  /**
   * Obtains the URL Encoded Cloud Logging LogName by reading the proto annotation of the {@link
   * AuditLogRequest.LogType}. If the proto annotation is missing, we default to {@code
   * LogType.UNSPECIFIED}.
   *
   * @param type logType
   * @return log name for the given log type
   */
  private String getLogNameFromLogType(LogType type) throws UnsupportedEncodingException {
    String logName =
        type.getValueDescriptor().getOptions().getExtension(AuditLogRequestProto.logName);
    return logName.isEmpty()
        ? LogType.UNSPECIFIED.name()
        : URLEncoder.encode(logName, StandardCharsets.UTF_8);
  }
}
