// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: echoer.proto

package echoer

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

// EchoerClient is the client API for Echoer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EchoerClient interface {
	Echo(ctx context.Context, opts ...grpc.CallOption) (Echoer_EchoClient, error)
}

type echoerClient struct {
	cc grpc.ClientConnInterface
}

func NewEchoerClient(cc grpc.ClientConnInterface) EchoerClient {
	return &echoerClient{cc}
}

func (c *echoerClient) Echo(ctx context.Context, opts ...grpc.CallOption) (Echoer_EchoClient, error) {
	stream, err := c.cc.NewStream(ctx, &Echoer_ServiceDesc.Streams[0], "/echoer.Echoer/Echo", opts...)
	if err != nil {
		return nil, err
	}
	x := &echoerEchoClient{stream}
	return x, nil
}

type Echoer_EchoClient interface {
	Send(*EchoRequest) error
	Recv() (*EchoResponse, error)
	grpc.ClientStream
}

type echoerEchoClient struct {
	grpc.ClientStream
}

func (x *echoerEchoClient) Send(m *EchoRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *echoerEchoClient) Recv() (*EchoResponse, error) {
	m := new(EchoResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EchoerServer is the server API for Echoer service.
// All implementations must embed UnimplementedEchoerServer
// for forward compatibility
type EchoerServer interface {
	Echo(Echoer_EchoServer) error
	mustEmbedUnimplementedEchoerServer()
}

// UnimplementedEchoerServer must be embedded to have forward compatible implementations.
type UnimplementedEchoerServer struct {
}

func (UnimplementedEchoerServer) Echo(Echoer_EchoServer) error {
	return status.Errorf(codes.Unimplemented, "method Echo not implemented")
}
func (UnimplementedEchoerServer) mustEmbedUnimplementedEchoerServer() {}

// UnsafeEchoerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EchoerServer will
// result in compilation errors.
type UnsafeEchoerServer interface {
	mustEmbedUnimplementedEchoerServer()
}

func RegisterEchoerServer(s grpc.ServiceRegistrar, srv EchoerServer) {
	s.RegisterService(&Echoer_ServiceDesc, srv)
}

func _Echoer_Echo_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EchoerServer).Echo(&echoerEchoServer{stream})
}

type Echoer_EchoServer interface {
	Send(*EchoResponse) error
	Recv() (*EchoRequest, error)
	grpc.ServerStream
}

type echoerEchoServer struct {
	grpc.ServerStream
}

func (x *echoerEchoServer) Send(m *EchoResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *echoerEchoServer) Recv() (*EchoRequest, error) {
	m := new(EchoRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Echoer_ServiceDesc is the grpc.ServiceDesc for Echoer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Echoer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "echoer.Echoer",
	HandlerType: (*EchoerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Echo",
			Handler:       _Echoer_Echo_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "echoer.proto",
}
