// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package talkerpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TalkerClient is the client API for Talker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TalkerClient interface {
	// Say hello with something OK to audit log in request/response.
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
	// Whisper with something sensitive (shouldn't be audit logged) in
	// request/response.
	Whisper(ctx context.Context, in *WhisperRequest, opts ...grpc.CallOption) (*WhisperResponse, error)
	// Say byte with something OK to audit log in request,
	// but the response is empty.
	Bye(ctx context.Context, in *ByeRequest, opts ...grpc.CallOption) (*ByeResponse, error)
}

type talkerClient struct {
	cc grpc.ClientConnInterface
}

func NewTalkerClient(cc grpc.ClientConnInterface) TalkerClient {
	return &talkerClient{cc}
}

func (c *talkerClient) Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	out := new(HelloResponse)
	err := c.cc.Invoke(ctx, "/abcxyz.test.Talker/Hello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *talkerClient) Whisper(ctx context.Context, in *WhisperRequest, opts ...grpc.CallOption) (*WhisperResponse, error) {
	out := new(WhisperResponse)
	err := c.cc.Invoke(ctx, "/abcxyz.test.Talker/Whisper", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *talkerClient) Bye(ctx context.Context, in *ByeRequest, opts ...grpc.CallOption) (*ByeResponse, error) {
	out := new(ByeResponse)
	err := c.cc.Invoke(ctx, "/abcxyz.test.Talker/Bye", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TalkerServer is the server API for Talker service.
// All implementations must embed UnimplementedTalkerServer
// for forward compatibility
type TalkerServer interface {
	// Say hello with something OK to audit log in request/response.
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	// Whisper with something sensitive (shouldn't be audit logged) in
	// request/response.
	Whisper(context.Context, *WhisperRequest) (*WhisperResponse, error)
	// Say byte with something OK to audit log in request,
	// but the response is empty.
	Bye(context.Context, *ByeRequest) (*ByeResponse, error)
	mustEmbedUnimplementedTalkerServer()
}

// UnimplementedTalkerServer must be embedded to have forward compatible implementations.
type UnimplementedTalkerServer struct {
}

func (UnimplementedTalkerServer) Hello(context.Context, *HelloRequest) (*HelloResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hello not implemented")
}
func (UnimplementedTalkerServer) Whisper(context.Context, *WhisperRequest) (*WhisperResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Whisper not implemented")
}
func (UnimplementedTalkerServer) Bye(context.Context, *ByeRequest) (*ByeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Bye not implemented")
}
func (UnimplementedTalkerServer) mustEmbedUnimplementedTalkerServer() {}

// UnsafeTalkerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TalkerServer will
// result in compilation errors.
type UnsafeTalkerServer interface {
	mustEmbedUnimplementedTalkerServer()
}

func RegisterTalkerServer(s grpc.ServiceRegistrar, srv TalkerServer) {
	s.RegisterService(&Talker_ServiceDesc, srv)
}

func _Talker_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TalkerServer).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/abcxyz.test.Talker/Hello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TalkerServer).Hello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Talker_Whisper_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WhisperRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TalkerServer).Whisper(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/abcxyz.test.Talker/Whisper",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TalkerServer).Whisper(ctx, req.(*WhisperRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Talker_Bye_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ByeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TalkerServer).Bye(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/abcxyz.test.Talker/Bye",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TalkerServer).Bye(ctx, req.(*ByeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Talker_ServiceDesc is the grpc.ServiceDesc for Talker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Talker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "abcxyz.test.Talker",
	HandlerType: (*TalkerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    _Talker_Hello_Handler,
		},
		{
			MethodName: "Whisper",
			Handler:    _Talker_Whisper_Handler,
		},
		{
			MethodName: "Bye",
			Handler:    _Talker_Bye_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "integration/protos/talker.proto",
}
