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

package abcxyz.lumberjack.test.talker;

import com.abcxyz.lumberjack.auditlogclient.AuditLogs;
import com.abcxyz.lumberjack.test.talker.FailOnFourRequest;
import com.abcxyz.lumberjack.test.talker.FailOnFourResponse;
import com.google.cloud.audit.AuditLog;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import lombok.extern.slf4j.Slf4j;

/** Server-side handler for client streaming. This one fails if it ever receives the value "4" */
@Slf4j
public class ServerFailOnFourObserver implements StreamObserver<FailOnFourRequest> {
  private final StreamObserver<FailOnFourResponse> responseStream;

  public ServerFailOnFourObserver(StreamObserver<FailOnFourResponse> responseStream) {
    this.responseStream = responseStream;
  }

  @Override
  public void onNext(FailOnFourRequest request) {
    AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
    auditLogBuilder.setResourceName(request.getTarget());
    if (request.getValue() == 4) {
      log.info("Got 4, throwing error");
      // onError(new StatusRuntimeException(Status.INVALID_ARGUMENT));
      throw new StatusRuntimeException(Status.INVALID_ARGUMENT);
    } else {
      log.info("Got {} which isn't 4.", request.getValue());
    }
  }

  @Override
  public void onError(Throwable t) {
    responseStream.onError(t);
  }

  @Override
  public void onCompleted() {
    FailOnFourResponse response =
        FailOnFourResponse.newBuilder().setMessage("Good job sending no 4s").build();
    responseStream.onNext(response);
    responseStream.onCompleted();
  }
}
