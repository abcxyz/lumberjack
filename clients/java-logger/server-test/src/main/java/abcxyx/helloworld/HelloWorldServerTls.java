package abcxyx.helloworld;

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

import abcxyx.helloworld.generated.GreeterGrpc;
import abcxyx.helloworld.generated.HelloReply;
import abcxyx.helloworld.generated.HelloRequest;
import com.abcxyz.lumberjack.auditlogclient.AuditLoggingServerInterceptor;
import com.abcxyz.lumberjack.auditlogclient.AuditLogs;
import com.abcxyz.lumberjack.auditlogclient.modules.AuditLoggingModule;
import com.google.cloud.audit.AuditLog;
import com.google.inject.Guice;
import com.google.inject.Injector;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.util.logging.Logger;
import lombok.RequiredArgsConstructor;

/** Server that manages startup/shutdown of a {@code Greeter} server with TLS enabled. */
@RequiredArgsConstructor
public class HelloWorldServerTls {
  private static final Logger logger = Logger.getLogger(HelloWorldServerTls.class.getName());

  private Server server;

  private final int port;

  private void start(AuditLoggingServerInterceptor interceptor) throws IOException {
    server =
        ServerBuilder.forPort(port)
            .addService(new GreeterImpl())
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
                HelloWorldServerTls.this.stop();
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

    final HelloWorldServerTls server =
        new HelloWorldServerTls(Integer.parseInt(System.getenv("PORT")));
    server.start(interceptor);
    server.blockUntilShutdown();
  }

  static class GreeterImpl extends GreeterGrpc.GreeterImplBase {

    @Override
    public void sayHello(HelloRequest req, StreamObserver<HelloReply> responseObserver) {
      HelloReply reply = HelloReply.newBuilder().setMessage("Hello " + req.getName()).build();

      AuditLog.Builder auditLogBuilder = AuditLogs.getBuilderFromContext();
      auditLogBuilder.setResourceName("MyResource");

      logger.info("replying");
      responseObserver.onNext(reply);
      responseObserver.onCompleted();
    }
  }
}
