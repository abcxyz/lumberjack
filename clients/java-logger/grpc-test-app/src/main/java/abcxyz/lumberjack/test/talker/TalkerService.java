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

import com.abcxyz.lumberjack.auditlogclient.AuditLoggingServerInterceptor;
import com.abcxyz.lumberjack.auditlogclient.AuditLogs;
import com.abcxyz.lumberjack.auditlogclient.modules.AuditLoggingModule;
import com.abcxyz.lumberjack.test.talker.AdditionRequest;
import com.abcxyz.lumberjack.test.talker.AdditionResponse;
import com.abcxyz.lumberjack.test.talker.FailRequest;
import com.abcxyz.lumberjack.test.talker.FailResponse;
import com.abcxyz.lumberjack.test.talker.FailOnFourRequest;
import com.abcxyz.lumberjack.test.talker.FailOnFourResponse;
import com.abcxyz.lumberjack.test.talker.FibonacciRequest;
import com.abcxyz.lumberjack.test.talker.FibonacciResponse;
import com.abcxyz.lumberjack.test.talker.HelloRequest;
import com.abcxyz.lumberjack.test.talker.HelloResponse;
import com.abcxyz.lumberjack.test.talker.TalkerGrpc;
import com.abcxyz.lumberjack.test.talker.WhisperRequest;
import com.abcxyz.lumberjack.test.talker.WhisperResponse;
import com.google.cloud.audit.AuditLog;
import com.google.inject.Guice;
import com.google.inject.Injector;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.logging.Logger;
import lombok.RequiredArgsConstructor;

/** Server that manages startup/shutdown of a {@code Talker} server with TLS enabled. */
@RequiredArgsConstructor
public class TalkerService {
  private static final Logger logger = Logger.getLogger(TalkerService.class.getName());
  private static final Map<Integer, Integer> fibonacciMemo = new HashMap<>();

  private Server server;

  private final int port;

  private void start(AuditLoggingServerInterceptor interceptor) throws IOException {
    server =
        ServerBuilder.forPort(port)
            .addService(new TalkerImpl())
            .intercept(interceptor)
            .build()
            .start();
    logger.info("Server started, listening on " + port);
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                // Use stderr here since the logger may have been reset by its JVM shutdown hook.
                System.err.println("*** shutting down gRPC server since JVM is shutting down");
                TalkerService.this.stop();
                System.err.println("*** server shut down");
              }
            });
  }

  private void stop() {
    if (server != null) {
      server.shutdown();
    }
  }

  /** Await termination on the main thread since the grpc library uses daemon threads. */
  private void blockUntilShutdown() throws InterruptedException {
    if (server != null) {
      server.awaitTermination();
    }
  }

  /** Main launches the server from the command line. */
  public static void main(String[] args) throws IOException, InterruptedException {
    Injector injector = Guice.createInjector(new AuditLoggingModule());
    AuditLoggingServerInterceptor interceptor =
        injector.getInstance(AuditLoggingServerInterceptor.class);

    final TalkerService server = new TalkerService(Integer.parseInt(System.getenv("PORT")));
    server.start(interceptor);
    server.blockUntilShutdown();
  }

  private static class TalkerImpl extends TalkerGrpc.TalkerImplBase {
    @Override
    public void hello(HelloRequest req, StreamObserver<HelloResponse> responseObserver) {
      HelloResponse reply =
          HelloResponse.newBuilder().setMessage("Hello " + req.getMessage()).build();

      AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
      auditLogBuilder.setResourceName(req.getTarget());

      logger.info("replying");
      responseObserver.onNext(reply);
      responseObserver.onCompleted();
    }

    @Override
    public void whisper(WhisperRequest req, StreamObserver<WhisperResponse> responseObserver) {
      WhisperResponse reply =
          WhisperResponse.newBuilder().setMessage("I'll keep that secret!").build();

      AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
      auditLogBuilder.setResourceName(req.getTarget());

      logger.info("replying");
      responseObserver.onNext(reply);
      responseObserver.onCompleted();
    }

    /**
     * This is a test API for server streaming. The client sends a request with how many places of
     * fibonacci numbers it wants, and then the server streams each number in order.
     *
     * example: 3 places -> 0, 1, 1
     */
    @Override
    public void fibonacci(
        FibonacciRequest request, StreamObserver<FibonacciResponse> responseObserver) {
      for (int i = 0; i < request.getPlaces(); i++) {
        int value = getFibonacciValueForPosition(i);
        FibonacciResponse response =
            FibonacciResponse.newBuilder().setPosition(i + 1).setValue(value).build();
        AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
        auditLogBuilder.setResourceName(request.getTarget());
        responseObserver.onNext(response);
      }
      responseObserver.onCompleted();
    }

    private int getFibonacciValueForPosition(int position) {
      if (position == 0) return 0;
      if (position == 1 || position == 2) return 1;
      if (fibonacciMemo.containsKey(position)) return fibonacciMemo.get(position);

      int value = getFibonacciValueForPosition(position - 1) + getFibonacciValueForPosition(position - 2);
      fibonacciMemo.put(position, value);
      return value;
    }

    /**
     * This is a test API for client streaming. The client opens a stream and can send any
     * number of numbers. The server adds up all those numbers, and when the stream is closed,
     * replies with the final sum of all the numbers.
     */
    @Override
    public StreamObserver<AdditionRequest> addition(
        StreamObserver<AdditionResponse> responseObserver) {
      return new ServerAdditionObserver(responseObserver);
    }

    /**
     * This is an api that always fails. It is intended to test the failure modes of our application.
     */
    @Override
    public void fail(FailRequest req, StreamObserver<FailResponse> responseObserver) {
      AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
      auditLogBuilder.setResourceName(req.getTarget());
      throw new StatusRuntimeException(Status.RESOURCE_EXHAUSTED);
    }

    /**
     * This fails if it receives the value "4". Intended for testing what happens on a failure mid-stream.
     */
    @Override
    public StreamObserver<FailOnFourRequest> failOnFour(
        StreamObserver<FailOnFourResponse> responseObserver) {
      return new ServerFailOnFourObserver(responseObserver);
    }
  }
}
