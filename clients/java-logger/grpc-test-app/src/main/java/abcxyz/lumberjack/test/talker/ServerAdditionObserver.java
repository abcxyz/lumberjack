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
import com.abcxyz.lumberjack.test.talker.AdditionRequest;
import com.abcxyz.lumberjack.test.talker.AdditionResponse;
import com.google.cloud.audit.AuditLog;
import io.grpc.stub.StreamObserver;

/** Server-side handler for client streaming. */
public class ServerAdditionObserver implements StreamObserver<AdditionRequest> {
  private int sum = 0;
  private final StreamObserver<AdditionResponse> responseStream;

  public ServerAdditionObserver(StreamObserver<AdditionResponse> responseStream) {
    this.responseStream = responseStream;
  }

  @Override
  public void onNext(AdditionRequest request) {
    AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
    auditLogBuilder.setResourceName(request.getTarget());
    sum += request.getAddend();
  }

  @Override
  public void onError(Throwable t) {
    // no-op
  }

  @Override
  public void onCompleted() {
    AdditionResponse response = AdditionResponse.newBuilder().setSum(sum).build();
    responseStream.onNext(response);
    responseStream.onCompleted();
  }
}
