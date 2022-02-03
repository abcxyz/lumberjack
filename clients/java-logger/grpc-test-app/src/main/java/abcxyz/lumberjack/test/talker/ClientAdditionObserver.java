package abcxyz.lumberjack.test.talker;

import com.abcxyz.lumberjack.test.talker.AdditionResponse;
import io.grpc.stub.StreamObserver;
import java.util.logging.Logger;

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
