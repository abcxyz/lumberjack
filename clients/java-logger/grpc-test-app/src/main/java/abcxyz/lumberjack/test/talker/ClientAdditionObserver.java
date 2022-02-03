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

import com.abcxyz.lumberjack.test.talker.AdditionResponse;
import io.grpc.stub.StreamObserver;
import java.util.logging.Logger;

/**
 * Client-side handler for client streaming.
 */
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
