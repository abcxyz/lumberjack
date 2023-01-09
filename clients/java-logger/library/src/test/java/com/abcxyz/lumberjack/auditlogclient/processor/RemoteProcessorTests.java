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

import static org.mockito.AdditionalAnswers.delegatesTo;
import static org.mockito.Mockito.any;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

import com.abcxyz.lumberjack.v1alpha1.AuditLogAgentGrpc;
import com.abcxyz.lumberjack.v1alpha1.AuditLogAgentGrpc.AuditLogAgentBlockingStub;
import com.abcxyz.lumberjack.v1alpha1.AuditLogRequest;
import com.abcxyz.lumberjack.v1alpha1.AuditLogResponse;
import io.grpc.ManagedChannel;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import io.grpc.stub.StreamObserver;
import io.grpc.testing.GrpcCleanupRule;
import java.util.concurrent.TimeUnit;
import org.junit.After;
import org.junit.Before;
import org.junit.Rule;
import org.junit.Test;

public class RemoteProcessorTests {
  @Rule public final GrpcCleanupRule grpcCleanup = new GrpcCleanupRule();

  private static final class FakeAuditLogAgentImpl extends AuditLogAgentGrpc.AuditLogAgentImplBase {
    @Override
    public void processLog(
        AuditLogRequest request, StreamObserver<AuditLogResponse> responseObserver) {
      AuditLogResponse response =
          AuditLogResponse.newBuilder()
              .setResult(AuditLogRequest.newBuilder().getDefaultInstanceForType())
              .build();
      responseObserver.onNext(response);
      responseObserver.onCompleted();
    }
  }

  private final AuditLogAgentGrpc.AuditLogAgentImplBase serviceImpl =
      mock(AuditLogAgentGrpc.AuditLogAgentImplBase.class, delegatesTo(new FakeAuditLogAgentImpl()));

  private RemoteProcessor remoteProcessor;
  private ManagedChannel channel;

  @Before
  public void setUp() throws Exception {
    String serverName = InProcessServerBuilder.generateName();
    grpcCleanup.register(
        InProcessServerBuilder.forName(serverName)
            .directExecutor()
            .addService(serviceImpl)
            .build()
            .start());
    channel =
        grpcCleanup.register(InProcessChannelBuilder.forName(serverName).directExecutor().build());
    AuditLogAgentBlockingStub blockingStub = AuditLogAgentGrpc.newBlockingStub(channel);
    remoteProcessor = new RemoteProcessor(blockingStub);
  }

  @After
  public void tearDown() throws Exception {
    channel.shutdownNow().awaitTermination(5, TimeUnit.SECONDS);
  }

  @Test
  public void invokesBlockingStubWithAuditLogRequest() {
    AuditLogRequest request = AuditLogRequest.getDefaultInstance();
    remoteProcessor.process(request);
    verify(serviceImpl).processLog(any(), any());
  }
}
