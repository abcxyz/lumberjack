package abcxyz.lumberjack.test.talker;

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
/*
 * Copyright 2015 The gRPC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import com.abcxyz.lumberjack.test.talker.AdditionRequest;
import com.abcxyz.lumberjack.test.talker.FailOnFourRequest;
import com.abcxyz.lumberjack.test.talker.FailRequest;
import com.abcxyz.lumberjack.test.talker.FailResponse;
import com.abcxyz.lumberjack.test.talker.FibonacciRequest;
import com.abcxyz.lumberjack.test.talker.HelloRequest;
import com.abcxyz.lumberjack.test.talker.HelloResponse;
import com.abcxyz.lumberjack.test.talker.TalkerGrpc;
import com.abcxyz.lumberjack.test.talker.WhisperRequest;
import com.abcxyz.lumberjack.test.talker.WhisperResponse;
import com.google.auth.oauth2.AccessToken;
import com.google.auth.oauth2.GoogleCredentials;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import io.grpc.auth.MoreCallCredentials;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.util.Calendar;
import java.util.UUID;
import java.util.concurrent.TimeUnit;
import lombok.extern.slf4j.Slf4j;

/** A simple client that requests a greeting from the {@link TalkerService} with TLS. */
@Slf4j
public class TalkerClient {
  private final TalkerGrpc.TalkerBlockingStub blockingStub;
  private final TalkerGrpc.TalkerStub clientStub;

  /** Construct client for accessing {@link TalkerService} using the existing channel. */
  public TalkerClient(ManagedChannel channel, GoogleCredentials credentials) throws IOException {
    blockingStub =
        TalkerGrpc.newBlockingStub(channel)
            .withCallCredentials(MoreCallCredentials.from(credentials));
    clientStub =
        TalkerGrpc.newStub(channel).withCallCredentials(MoreCallCredentials.from(credentials));
  }

  /** Say hello to server. */
  public void greet(String name, UUID target) {
    log.info("Will try to greet " + name + " ...");
    HelloRequest request =
        HelloRequest.newBuilder().setMessage(name).setTarget(target.toString()).build();
    HelloResponse response;
    try {
      response = blockingStub.hello(request);
    } catch (StatusRuntimeException e) {
      log.info("RPC failed: " + e.getStatus());
      throw e;
    }
    log.info("Greeting: " + response.getMessage());
  }

  /** Whisper secrets to server. */
  public void whisper(String secret, UUID target) {
    log.info("Will try to whisper " + secret + " ...");
    WhisperRequest request =
        WhisperRequest.newBuilder().setMessage(secret).setTarget(target.toString()).build();
    WhisperResponse response;
    try {
      response = blockingStub.whisper(request);
    } catch (StatusRuntimeException e) {
      log.info("RPC failed: " + e.getStatus());
      throw e;
    }
    log.info("Greeting: " + response.getMessage());
  }

  /** Test failure. */
  public void fail(String message, UUID target) {
    log.info("Sending message " + message + " ...");
    FailRequest request =
        FailRequest.newBuilder().setMessage(message).setTarget(target.toString()).build();
    FailResponse response;
    try {
      response = blockingStub.fail(request);
      log.info("Did not receive error: " + response.getMessage());
      throw new IllegalStateException("Did not receive error from fail api");
    } catch (StatusRuntimeException e) {
      log.info("RPC failed: " + e.getStatus());
    } catch (Exception e) {
      log.info("Got other failure " + e.getMessage());
    }
  }

  public void fibonacci(int places, UUID target) {
    FibonacciRequest request =
        FibonacciRequest.newBuilder().setPlaces(places).setTarget(target.toString()).build();

    try {
      log.info("Fibonacci sequence for places " + places);
      blockingStub
          .fibonacci(request)
          .forEachRemaining(
              fibonacciResponse -> {
                log.info(
                    "Position: "
                        + fibonacciResponse.getPosition()
                        + " Value: "
                        + fibonacciResponse.getValue());
              });
    } catch (StatusRuntimeException e) {
      log.info("RPC failed: " + e.getStatus());
      throw e;
    }
  }

  public void addition(int max, UUID target) {
    StreamObserver<AdditionRequest> requestObserver =
        clientStub.addition(new ClientAdditionObserver());

    for (int i = 1; i <= max; i++) {
      log.info("Adding: " + i);
      AdditionRequest request =
          AdditionRequest.newBuilder().setAddend(i).setTarget(target.toString()).build();
      requestObserver.onNext(request);
    }

    requestObserver.onCompleted();
  }

  public void failOnFour(int max, UUID target) {
    StreamObserver<FailOnFourRequest> requestObserver =
        clientStub.failOnFour(new ClientFailOnFourObserver());

    for (int i = 1; i <= max; i++) {
      log.info("Sending: " + i);
      FailOnFourRequest request =
          FailOnFourRequest.newBuilder().setValue(1).setTarget(target.toString()).build();
      requestObserver.onNext(request);
    }

    requestObserver.onCompleted();
  }

  /**
   * Greet server. If provided, the first element of {@code args} is the name to use in the
   * greeting. First element can either be a list ( in format '["a", "b"]') or a singular host
   */
  public static void main(String[] args) throws Exception {
    // this turns an array string into an array. e.g. "["a", "b"]" -> ["a","b"]
    String hostList = args[0];
    String[] hosts =
        hostList
            .replaceAll("\\[", "")
            .replaceAll("\\]", "")
            .replace("https://", "")
            .replaceAll("\\s", "")
            .replaceAll("\"", "")
            .split(",");

    int port = Integer.parseInt(args[1]);

    GoogleCredentials credentials;
    if (args.length >= 3) {
      // token explicitly added to args, use that.
      log.info("Using explicit token");
      String token = args[2];
      Calendar currentTime = Calendar.getInstance();
      currentTime.add(Calendar.MINUTE, 15);
      credentials = GoogleCredentials.create(new AccessToken(token, currentTime.getTime()));
    } else {
      log.info("Attempting to use default credentials");
      // try to use the application default credentials if no token is specified.
      credentials = GoogleCredentials.getApplicationDefault();
    }

    for (String host : hosts) {
      ManagedChannel channel = ManagedChannelBuilder.forAddress(host, port).build();

      try {
        TalkerClient client = new TalkerClient(channel, credentials);
        UUID target = UUID.randomUUID();
        client.greet(host, target);
        client.whisper("This is a secret! Don't audit log this string", target);
        client.fibonacci(5, target);
        client.addition(3, target);
        client.fail("This message should result in failure", target);
        client.failOnFour(5, target);
        // Sleep and wait for response to addition. Blocking stub doesn't support client streaming.
        TimeUnit.SECONDS.sleep(5);
      } finally {
        channel.shutdownNow().awaitTermination(5, TimeUnit.SECONDS);
      }
    }
  }
}
