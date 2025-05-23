// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.19.6
// source: proto/product.proto

package productpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ProductService_GetProductStock_FullMethodName    = "/productpb.ProductService/GetProductStock"
	ProductService_UpdateProductStock_FullMethodName = "/productpb.ProductService/UpdateProductStock"
)

// ProductServiceClient is the client API for ProductService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Определяем сервис ProductService
type ProductServiceClient interface {
	// Метод для получения информации о stock
	GetProductStock(ctx context.Context, in *ProductStockRequest, opts ...grpc.CallOption) (*ProductStockResponse, error)
	// Метод для обновления количества продуктов
	UpdateProductStock(ctx context.Context, in *UpdateProductStockRequest, opts ...grpc.CallOption) (*UpdateProductStockResponse, error)
}

type productServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProductServiceClient(cc grpc.ClientConnInterface) ProductServiceClient {
	return &productServiceClient{cc}
}

func (c *productServiceClient) GetProductStock(ctx context.Context, in *ProductStockRequest, opts ...grpc.CallOption) (*ProductStockResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ProductStockResponse)
	err := c.cc.Invoke(ctx, ProductService_GetProductStock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *productServiceClient) UpdateProductStock(ctx context.Context, in *UpdateProductStockRequest, opts ...grpc.CallOption) (*UpdateProductStockResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateProductStockResponse)
	err := c.cc.Invoke(ctx, ProductService_UpdateProductStock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProductServiceServer is the server API for ProductService service.
// All implementations must embed UnimplementedProductServiceServer
// for forward compatibility.
//
// Определяем сервис ProductService
type ProductServiceServer interface {
	// Метод для получения информации о stock
	GetProductStock(context.Context, *ProductStockRequest) (*ProductStockResponse, error)
	// Метод для обновления количества продуктов
	UpdateProductStock(context.Context, *UpdateProductStockRequest) (*UpdateProductStockResponse, error)
	mustEmbedUnimplementedProductServiceServer()
}

// UnimplementedProductServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedProductServiceServer struct{}

func (UnimplementedProductServiceServer) GetProductStock(context.Context, *ProductStockRequest) (*ProductStockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProductStock not implemented")
}
func (UnimplementedProductServiceServer) UpdateProductStock(context.Context, *UpdateProductStockRequest) (*UpdateProductStockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProductStock not implemented")
}
func (UnimplementedProductServiceServer) mustEmbedUnimplementedProductServiceServer() {}
func (UnimplementedProductServiceServer) testEmbeddedByValue()                        {}

// UnsafeProductServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProductServiceServer will
// result in compilation errors.
type UnsafeProductServiceServer interface {
	mustEmbedUnimplementedProductServiceServer()
}

func RegisterProductServiceServer(s grpc.ServiceRegistrar, srv ProductServiceServer) {
	// If the following call pancis, it indicates UnimplementedProductServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ProductService_ServiceDesc, srv)
}

func _ProductService_GetProductStock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProductStockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).GetProductStock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProductService_GetProductStock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServiceServer).GetProductStock(ctx, req.(*ProductStockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductService_UpdateProductStock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateProductStockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductServiceServer).UpdateProductStock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ProductService_UpdateProductStock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductServiceServer).UpdateProductStock(ctx, req.(*UpdateProductStockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ProductService_ServiceDesc is the grpc.ServiceDesc for ProductService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ProductService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "productpb.ProductService",
	HandlerType: (*ProductServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetProductStock",
			Handler:    _ProductService_GetProductStock_Handler,
		},
		{
			MethodName: "UpdateProductStock",
			Handler:    _ProductService_UpdateProductStock_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/product.proto",
}
