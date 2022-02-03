package abcxyz.lumberjack.test.talker;

import com.abcxyz.lumberjack.auditlogclient.AuditLogs;
import com.abcxyz.lumberjack.test.talker.AdditionRequest;
import com.abcxyz.lumberjack.test.talker.AdditionResponse;
import com.google.cloud.audit.AuditLog;
import io.grpc.stub.StreamObserver;

/**
 * Server-side handler for client streaming.
 */
public class ServerAdditionObserver implements StreamObserver<AdditionRequest> {
  private int sum = 0;
  private final StreamObserver<AdditionResponse> responseStream;

  public ServerAdditionObserver(StreamObserver<AdditionResponse> responseStream) {
    this.responseStream = responseStream;
  }

  @Override
  public void onNext(AdditionRequest request) {
    sum = sum + request.getAddend();
  }

  @Override
  public void onError(Throwable t) {
    // no-op
  }

  @Override
  public void onCompleted() {
    AdditionResponse response = AdditionResponse.newBuilder().setSum(sum).build();
    AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
    // TODO: what should we set as resource names for client streaming?
    auditLogBuilder.setResourceName("Placeholder");
    responseStream.onNext(response);
    responseStream.onCompleted();
  }
}
