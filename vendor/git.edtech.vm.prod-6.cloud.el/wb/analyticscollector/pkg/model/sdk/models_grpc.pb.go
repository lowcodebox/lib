// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.6
// source: models.proto

package __

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

// CollectorClient is the client API for Collector service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CollectorClient interface {
	Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error)
	Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error)
}

type collectorClient struct {
	cc grpc.ClientConnInterface
}

func NewCollectorClient(cc grpc.ClientConnInterface) CollectorClient {
	return &collectorClient{cc}
}

func (c *collectorClient) Set(ctx context.Context, in *SetRequest, opts ...grpc.CallOption) (*SetResponse, error) {
	out := new(SetResponse)
	err := c.cc.Invoke(ctx, "/analytic_collector.Collector/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectorClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	out := new(SearchResponse)
	err := c.cc.Invoke(ctx, "/analytic_collector.Collector/Search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectorClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error) {
	out := new(QueryResponse)
	err := c.cc.Invoke(ctx, "/analytic_collector.Collector/Query", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CollectorServer is the server API for Collector service.
// All implementations should embed UnimplementedCollectorServer
// for forward compatibility
type CollectorServer interface {
	Set(context.Context, *SetRequest) (*SetResponse, error)
	Search(context.Context, *SearchRequest) (*SearchResponse, error)
	Query(context.Context, *QueryRequest) (*QueryResponse, error)
}

// UnimplementedCollectorServer should be embedded to have forward compatible implementations.
type UnimplementedCollectorServer struct {
}

func (UnimplementedCollectorServer) Set(context.Context, *SetRequest) (*SetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (UnimplementedCollectorServer) Search(context.Context, *SearchRequest) (*SearchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedCollectorServer) Query(context.Context, *QueryRequest) (*QueryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}

// UnsafeCollectorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CollectorServer will
// result in compilation errors.
type UnsafeCollectorServer interface {
	mustEmbedUnimplementedCollectorServer()
}

func RegisterCollectorServer(s grpc.ServiceRegistrar, srv CollectorServer) {
	s.RegisterService(&Collector_ServiceDesc, srv)
}

func _Collector_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectorServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/analytic_collector.Collector/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectorServer).Set(ctx, req.(*SetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Collector_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectorServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/analytic_collector.Collector/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectorServer).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Collector_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectorServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/analytic_collector.Collector/Query",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectorServer).Query(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Collector_ServiceDesc is the grpc.ServiceDesc for Collector service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Collector_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "analytic_collector.Collector",
	HandlerType: (*CollectorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Set",
			Handler:    _Collector_Set_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _Collector_Search_Handler,
		},
		{
			MethodName: "Query",
			Handler:    _Collector_Query_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "models.proto",
}
