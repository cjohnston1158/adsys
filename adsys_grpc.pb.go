// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package adsys

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// ServiceClient is the client API for Service service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ServiceClient interface {
	Cat(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Service_CatClient, error)
	Version(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Service_VersionClient, error)
	Stop(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (Service_StopClient, error)
}

type serviceClient struct {
	cc grpc.ClientConnInterface
}

func NewServiceClient(cc grpc.ClientConnInterface) ServiceClient {
	return &serviceClient{cc}
}

func (c *serviceClient) Cat(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Service_CatClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Service_serviceDesc.Streams[0], "/service/Cat", opts...)
	if err != nil {
		return nil, err
	}
	x := &serviceCatClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Service_CatClient interface {
	Recv() (*StringResponse, error)
	grpc.ClientStream
}

type serviceCatClient struct {
	grpc.ClientStream
}

func (x *serviceCatClient) Recv() (*StringResponse, error) {
	m := new(StringResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *serviceClient) Version(ctx context.Context, in *Empty, opts ...grpc.CallOption) (Service_VersionClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Service_serviceDesc.Streams[1], "/service/Version", opts...)
	if err != nil {
		return nil, err
	}
	x := &serviceVersionClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Service_VersionClient interface {
	Recv() (*VersionResponse, error)
	grpc.ClientStream
}

type serviceVersionClient struct {
	grpc.ClientStream
}

func (x *serviceVersionClient) Recv() (*VersionResponse, error) {
	m := new(VersionResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *serviceClient) Stop(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (Service_StopClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Service_serviceDesc.Streams[2], "/service/Stop", opts...)
	if err != nil {
		return nil, err
	}
	x := &serviceStopClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Service_StopClient interface {
	Recv() (*Empty, error)
	grpc.ClientStream
}

type serviceStopClient struct {
	grpc.ClientStream
}

func (x *serviceStopClient) Recv() (*Empty, error) {
	m := new(Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ServiceServer is the server API for Service service.
// All implementations must embed UnimplementedServiceServer
// for forward compatibility
type ServiceServer interface {
	Cat(*Empty, Service_CatServer) error
	Version(*Empty, Service_VersionServer) error
	Stop(*StopRequest, Service_StopServer) error
	mustEmbedUnimplementedServiceServer()
}

// UnimplementedServiceServer must be embedded to have forward compatible implementations.
type UnimplementedServiceServer struct {
}

func (UnimplementedServiceServer) Cat(*Empty, Service_CatServer) error {
	return status.Errorf(codes.Unimplemented, "method Cat not implemented")
}
func (UnimplementedServiceServer) Version(*Empty, Service_VersionServer) error {
	return status.Errorf(codes.Unimplemented, "method Version not implemented")
}
func (UnimplementedServiceServer) Stop(*StopRequest, Service_StopServer) error {
	return status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedServiceServer) mustEmbedUnimplementedServiceServer() {}

// UnsafeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ServiceServer will
// result in compilation errors.
type UnsafeServiceServer interface {
	mustEmbedUnimplementedServiceServer()
}

func RegisterServiceServer(s grpc.ServiceRegistrar, srv ServiceServer) {
	s.RegisterService(&_Service_serviceDesc, srv)
}

func _Service_Cat_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ServiceServer).Cat(m, &serviceCatServer{stream})
}

type Service_CatServer interface {
	Send(*StringResponse) error
	grpc.ServerStream
}

type serviceCatServer struct {
	grpc.ServerStream
}

func (x *serviceCatServer) Send(m *StringResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Service_Version_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ServiceServer).Version(m, &serviceVersionServer{stream})
}

type Service_VersionServer interface {
	Send(*VersionResponse) error
	grpc.ServerStream
}

type serviceVersionServer struct {
	grpc.ServerStream
}

func (x *serviceVersionServer) Send(m *VersionResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Service_Stop_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StopRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ServiceServer).Stop(m, &serviceStopServer{stream})
}

type Service_StopServer interface {
	Send(*Empty) error
	grpc.ServerStream
}

type serviceStopServer struct {
	grpc.ServerStream
}

func (x *serviceStopServer) Send(m *Empty) error {
	return x.ServerStream.SendMsg(m)
}

var _Service_serviceDesc = grpc.ServiceDesc{
	ServiceName: "service",
	HandlerType: (*ServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Cat",
			Handler:       _Service_Cat_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Version",
			Handler:       _Service_Version_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Stop",
			Handler:       _Service_Stop_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "adsys.proto",
}
