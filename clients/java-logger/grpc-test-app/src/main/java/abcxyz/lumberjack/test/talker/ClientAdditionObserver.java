package abcxyz.lumberjack.test.talker;

import java.util.logging.Logger;
import com.abcxyz.lumberjack.test.talker.AdditionResponse;
import io.grpc.stub.StreamObserver;

public class ClientAdditionObserver implements StreamObserver<AdditionResponse> {
  private static final Logger logger = Logger.getLogger(TalkerClient.class.getName());

  @Override
  public void onNext(AdditionResponse response) {
    logger.info("Sum was " + response.getSum());
  }

  @Override
  public void onError(Throwable throwable) {
    // no-op
  }

  @Override
  public void onCompleted() {
    // no-op
  }
}
