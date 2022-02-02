package com.abcxyz.lumberjack.test.talker;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * A gRPC service definition used for lumberjack integration test.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.42.1)",
    comments = "Source: talker.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class TalkerGrpc {

  private TalkerGrpc() {}

  public static final String SERVICE_NAME = "abcxyz.test.Talker";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.HelloRequest,
      com.abcxyz.lumberjack.test.talker.HelloResponse> getHelloMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Hello",
      requestType = com.abcxyz.lumberjack.test.talker.HelloRequest.class,
      responseType = com.abcxyz.lumberjack.test.talker.HelloResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.HelloRequest,
      com.abcxyz.lumberjack.test.talker.HelloResponse> getHelloMethod() {
    io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.HelloRequest, com.abcxyz.lumberjack.test.talker.HelloResponse> getHelloMethod;
    if ((getHelloMethod = TalkerGrpc.getHelloMethod) == null) {
      synchronized (TalkerGrpc.class) {
        if ((getHelloMethod = TalkerGrpc.getHelloMethod) == null) {
          TalkerGrpc.getHelloMethod = getHelloMethod =
              io.grpc.MethodDescriptor.<com.abcxyz.lumberjack.test.talker.HelloRequest, com.abcxyz.lumberjack.test.talker.HelloResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Hello"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.HelloRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.HelloResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TalkerMethodDescriptorSupplier("Hello"))
              .build();
        }
      }
    }
    return getHelloMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.WhisperRequest,
      com.abcxyz.lumberjack.test.talker.WhisperResponse> getWhisperMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Whisper",
      requestType = com.abcxyz.lumberjack.test.talker.WhisperRequest.class,
      responseType = com.abcxyz.lumberjack.test.talker.WhisperResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.WhisperRequest,
      com.abcxyz.lumberjack.test.talker.WhisperResponse> getWhisperMethod() {
    io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.WhisperRequest, com.abcxyz.lumberjack.test.talker.WhisperResponse> getWhisperMethod;
    if ((getWhisperMethod = TalkerGrpc.getWhisperMethod) == null) {
      synchronized (TalkerGrpc.class) {
        if ((getWhisperMethod = TalkerGrpc.getWhisperMethod) == null) {
          TalkerGrpc.getWhisperMethod = getWhisperMethod =
              io.grpc.MethodDescriptor.<com.abcxyz.lumberjack.test.talker.WhisperRequest, com.abcxyz.lumberjack.test.talker.WhisperResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Whisper"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.WhisperRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.WhisperResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TalkerMethodDescriptorSupplier("Whisper"))
              .build();
        }
      }
    }
    return getWhisperMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.ByeRequest,
      com.abcxyz.lumberjack.test.talker.ByeResponse> getByeMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Bye",
      requestType = com.abcxyz.lumberjack.test.talker.ByeRequest.class,
      responseType = com.abcxyz.lumberjack.test.talker.ByeResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.ByeRequest,
      com.abcxyz.lumberjack.test.talker.ByeResponse> getByeMethod() {
    io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.ByeRequest, com.abcxyz.lumberjack.test.talker.ByeResponse> getByeMethod;
    if ((getByeMethod = TalkerGrpc.getByeMethod) == null) {
      synchronized (TalkerGrpc.class) {
        if ((getByeMethod = TalkerGrpc.getByeMethod) == null) {
          TalkerGrpc.getByeMethod = getByeMethod =
              io.grpc.MethodDescriptor.<com.abcxyz.lumberjack.test.talker.ByeRequest, com.abcxyz.lumberjack.test.talker.ByeResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Bye"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.ByeRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.ByeResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TalkerMethodDescriptorSupplier("Bye"))
              .build();
        }
      }
    }
    return getByeMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.FibonacciRequest,
      com.abcxyz.lumberjack.test.talker.FibonacciResponse> getFibonacciMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Fibonacci",
      requestType = com.abcxyz.lumberjack.test.talker.FibonacciRequest.class,
      responseType = com.abcxyz.lumberjack.test.talker.FibonacciResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.FibonacciRequest,
      com.abcxyz.lumberjack.test.talker.FibonacciResponse> getFibonacciMethod() {
    io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.FibonacciRequest, com.abcxyz.lumberjack.test.talker.FibonacciResponse> getFibonacciMethod;
    if ((getFibonacciMethod = TalkerGrpc.getFibonacciMethod) == null) {
      synchronized (TalkerGrpc.class) {
        if ((getFibonacciMethod = TalkerGrpc.getFibonacciMethod) == null) {
          TalkerGrpc.getFibonacciMethod = getFibonacciMethod =
              io.grpc.MethodDescriptor.<com.abcxyz.lumberjack.test.talker.FibonacciRequest, com.abcxyz.lumberjack.test.talker.FibonacciResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Fibonacci"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.FibonacciRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.FibonacciResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TalkerMethodDescriptorSupplier("Fibonacci"))
              .build();
        }
      }
    }
    return getFibonacciMethod;
  }

  private static volatile io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.AdditionRequest,
      com.abcxyz.lumberjack.test.talker.AdditionResponse> getAdditionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Addition",
      requestType = com.abcxyz.lumberjack.test.talker.AdditionRequest.class,
      responseType = com.abcxyz.lumberjack.test.talker.AdditionResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.CLIENT_STREAMING)
  public static io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.AdditionRequest,
      com.abcxyz.lumberjack.test.talker.AdditionResponse> getAdditionMethod() {
    io.grpc.MethodDescriptor<com.abcxyz.lumberjack.test.talker.AdditionRequest, com.abcxyz.lumberjack.test.talker.AdditionResponse> getAdditionMethod;
    if ((getAdditionMethod = TalkerGrpc.getAdditionMethod) == null) {
      synchronized (TalkerGrpc.class) {
        if ((getAdditionMethod = TalkerGrpc.getAdditionMethod) == null) {
          TalkerGrpc.getAdditionMethod = getAdditionMethod =
              io.grpc.MethodDescriptor.<com.abcxyz.lumberjack.test.talker.AdditionRequest, com.abcxyz.lumberjack.test.talker.AdditionResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.CLIENT_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Addition"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.AdditionRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.abcxyz.lumberjack.test.talker.AdditionResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TalkerMethodDescriptorSupplier("Addition"))
              .build();
        }
      }
    }
    return getAdditionMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static TalkerStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<TalkerStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<TalkerStub>() {
        @java.lang.Override
        public TalkerStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new TalkerStub(channel, callOptions);
        }
      };
    return TalkerStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static TalkerBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<TalkerBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<TalkerBlockingStub>() {
        @java.lang.Override
        public TalkerBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new TalkerBlockingStub(channel, callOptions);
        }
      };
    return TalkerBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static TalkerFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<TalkerFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<TalkerFutureStub>() {
        @java.lang.Override
        public TalkerFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new TalkerFutureStub(channel, callOptions);
        }
      };
    return TalkerFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * A gRPC service definition used for lumberjack integration test.
   * </pre>
   */
  public static abstract class TalkerImplBase implements io.grpc.BindableService {

    /**
     * <pre>
     * Say hello with something OK to audit log in request/response.
     * </pre>
     */
    public void hello(com.abcxyz.lumberjack.test.talker.HelloRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.HelloResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getHelloMethod(), responseObserver);
    }

    /**
     * <pre>
     * Whisper with something sensitive (shouldn't be audit logged) in
     * request/response.
     * </pre>
     */
    public void whisper(com.abcxyz.lumberjack.test.talker.WhisperRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.WhisperResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getWhisperMethod(), responseObserver);
    }

    /**
     * <pre>
     * Say bye with something OK to audit log in request,
     * but we don't care the response at all.
     * </pre>
     */
    public void bye(com.abcxyz.lumberjack.test.talker.ByeRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.ByeResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getByeMethod(), responseObserver);
    }

    /**
     */
    public void fibonacci(com.abcxyz.lumberjack.test.talker.FibonacciRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.FibonacciResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getFibonacciMethod(), responseObserver);
    }

    /**
     */
    public io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.AdditionRequest> addition(
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.AdditionResponse> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getAdditionMethod(), responseObserver);
    }

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
          .addMethod(
            getHelloMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                com.abcxyz.lumberjack.test.talker.HelloRequest,
                com.abcxyz.lumberjack.test.talker.HelloResponse>(
                  this, METHODID_HELLO)))
          .addMethod(
            getWhisperMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                com.abcxyz.lumberjack.test.talker.WhisperRequest,
                com.abcxyz.lumberjack.test.talker.WhisperResponse>(
                  this, METHODID_WHISPER)))
          .addMethod(
            getByeMethod(),
            io.grpc.stub.ServerCalls.asyncUnaryCall(
              new MethodHandlers<
                com.abcxyz.lumberjack.test.talker.ByeRequest,
                com.abcxyz.lumberjack.test.talker.ByeResponse>(
                  this, METHODID_BYE)))
          .addMethod(
            getFibonacciMethod(),
            io.grpc.stub.ServerCalls.asyncServerStreamingCall(
              new MethodHandlers<
                com.abcxyz.lumberjack.test.talker.FibonacciRequest,
                com.abcxyz.lumberjack.test.talker.FibonacciResponse>(
                  this, METHODID_FIBONACCI)))
          .addMethod(
            getAdditionMethod(),
            io.grpc.stub.ServerCalls.asyncClientStreamingCall(
              new MethodHandlers<
                com.abcxyz.lumberjack.test.talker.AdditionRequest,
                com.abcxyz.lumberjack.test.talker.AdditionResponse>(
                  this, METHODID_ADDITION)))
          .build();
    }
  }

  /**
   * <pre>
   * A gRPC service definition used for lumberjack integration test.
   * </pre>
   */
  public static final class TalkerStub extends io.grpc.stub.AbstractAsyncStub<TalkerStub> {
    private TalkerStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected TalkerStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new TalkerStub(channel, callOptions);
    }

    /**
     * <pre>
     * Say hello with something OK to audit log in request/response.
     * </pre>
     */
    public void hello(com.abcxyz.lumberjack.test.talker.HelloRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.HelloResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getHelloMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Whisper with something sensitive (shouldn't be audit logged) in
     * request/response.
     * </pre>
     */
    public void whisper(com.abcxyz.lumberjack.test.talker.WhisperRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.WhisperResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getWhisperMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Say bye with something OK to audit log in request,
     * but we don't care the response at all.
     * </pre>
     */
    public void bye(com.abcxyz.lumberjack.test.talker.ByeRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.ByeResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getByeMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void fibonacci(com.abcxyz.lumberjack.test.talker.FibonacciRequest request,
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.FibonacciResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getFibonacciMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.AdditionRequest> addition(
        io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.AdditionResponse> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncClientStreamingCall(
          getChannel().newCall(getAdditionMethod(), getCallOptions()), responseObserver);
    }
  }

  /**
   * <pre>
   * A gRPC service definition used for lumberjack integration test.
   * </pre>
   */
  public static final class TalkerBlockingStub extends io.grpc.stub.AbstractBlockingStub<TalkerBlockingStub> {
    private TalkerBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected TalkerBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new TalkerBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Say hello with something OK to audit log in request/response.
     * </pre>
     */
    public com.abcxyz.lumberjack.test.talker.HelloResponse hello(com.abcxyz.lumberjack.test.talker.HelloRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getHelloMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Whisper with something sensitive (shouldn't be audit logged) in
     * request/response.
     * </pre>
     */
    public com.abcxyz.lumberjack.test.talker.WhisperResponse whisper(com.abcxyz.lumberjack.test.talker.WhisperRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getWhisperMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Say bye with something OK to audit log in request,
     * but we don't care the response at all.
     * </pre>
     */
    public com.abcxyz.lumberjack.test.talker.ByeResponse bye(com.abcxyz.lumberjack.test.talker.ByeRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getByeMethod(), getCallOptions(), request);
    }

    /**
     */
    public java.util.Iterator<com.abcxyz.lumberjack.test.talker.FibonacciResponse> fibonacci(
        com.abcxyz.lumberjack.test.talker.FibonacciRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getFibonacciMethod(), getCallOptions(), request);
    }
  }

  /**
   * <pre>
   * A gRPC service definition used for lumberjack integration test.
   * </pre>
   */
  public static final class TalkerFutureStub extends io.grpc.stub.AbstractFutureStub<TalkerFutureStub> {
    private TalkerFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected TalkerFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new TalkerFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Say hello with something OK to audit log in request/response.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<com.abcxyz.lumberjack.test.talker.HelloResponse> hello(
        com.abcxyz.lumberjack.test.talker.HelloRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getHelloMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Whisper with something sensitive (shouldn't be audit logged) in
     * request/response.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<com.abcxyz.lumberjack.test.talker.WhisperResponse> whisper(
        com.abcxyz.lumberjack.test.talker.WhisperRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getWhisperMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Say bye with something OK to audit log in request,
     * but we don't care the response at all.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<com.abcxyz.lumberjack.test.talker.ByeResponse> bye(
        com.abcxyz.lumberjack.test.talker.ByeRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getByeMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_HELLO = 0;
  private static final int METHODID_WHISPER = 1;
  private static final int METHODID_BYE = 2;
  private static final int METHODID_FIBONACCI = 3;
  private static final int METHODID_ADDITION = 4;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final TalkerImplBase serviceImpl;
    private final int methodId;

    MethodHandlers(TalkerImplBase serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_HELLO:
          serviceImpl.hello((com.abcxyz.lumberjack.test.talker.HelloRequest) request,
              (io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.HelloResponse>) responseObserver);
          break;
        case METHODID_WHISPER:
          serviceImpl.whisper((com.abcxyz.lumberjack.test.talker.WhisperRequest) request,
              (io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.WhisperResponse>) responseObserver);
          break;
        case METHODID_BYE:
          serviceImpl.bye((com.abcxyz.lumberjack.test.talker.ByeRequest) request,
              (io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.ByeResponse>) responseObserver);
          break;
        case METHODID_FIBONACCI:
          serviceImpl.fibonacci((com.abcxyz.lumberjack.test.talker.FibonacciRequest) request,
              (io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.FibonacciResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_ADDITION:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.addition(
              (io.grpc.stub.StreamObserver<com.abcxyz.lumberjack.test.talker.AdditionResponse>) responseObserver);
        default:
          throw new AssertionError();
      }
    }
  }

  private static abstract class TalkerBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    TalkerBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.abcxyz.lumberjack.test.talker.TalkerProto.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("Talker");
    }
  }

  private static final class TalkerFileDescriptorSupplier
      extends TalkerBaseDescriptorSupplier {
    TalkerFileDescriptorSupplier() {}
  }

  private static final class TalkerMethodDescriptorSupplier
      extends TalkerBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final String methodName;

    TalkerMethodDescriptorSupplier(String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (TalkerGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new TalkerFileDescriptorSupplier())
              .addMethod(getHelloMethod())
              .addMethod(getWhisperMethod())
              .addMethod(getByeMethod())
              .addMethod(getFibonacciMethod())
              .addMethod(getAdditionMethod())
              .build();
        }
      }
    }
    return result;
  }
}
