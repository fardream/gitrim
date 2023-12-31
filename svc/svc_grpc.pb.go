// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.9
// source: svc.proto

package svc

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

const (
	GiTrim_InitRepoSync_FullMethodName            = "/gitrim.svc.GiTrim/InitRepoSync"
	GiTrim_SyncToSubRepo_FullMethodName           = "/gitrim.svc.GiTrim/SyncToSubRepo"
	GiTrim_CommitsFromSubRepo_FullMethodName      = "/gitrim.svc.GiTrim/CommitsFromSubRepo"
	GiTrim_CheckRepoSyncUpToDate_FullMethodName   = "/gitrim.svc.GiTrim/CheckRepoSyncUpToDate"
	GiTrim_CheckCommitsFromSubRepo_FullMethodName = "/gitrim.svc.GiTrim/CheckCommitsFromSubRepo"
	GiTrim_GetRepoSync_FullMethodName             = "/gitrim.svc.GiTrim/GetRepoSync"
)

// GiTrimClient is the client API for GiTrim service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GiTrimClient interface {
	// InitRepoSync setup the sync-ing between two repos.
	//
	// The ID of the created repo will be sha256 hash of the following string
	//
	//	(from remote name)-(from owner name)-(from repo name)-(from branch)-(to
	//	remote name)-(to owner name)-(to repo name)-(to branch)
	//
	// The operation will also generate a secret for git webhooks, the secret is
	// generated from the id + 16 byte long salt.
	InitRepoSync(ctx context.Context, in *InitRepoSyncRequest, opts ...grpc.CallOption) (*InitRepoSyncResponse, error)
	// SyncToSubRepo syncs from the original repo to the sub repo.
	// The request can be set to force.
	SyncToSubRepo(ctx context.Context, in *SyncToSubRepoRequest, opts ...grpc.CallOption) (*SyncToSubRepoResponse, error)
	// CommitsFromSubRepo tries to sends a series of commits from a sub repo to
	// the original repo.
	//
	// The commits will be rejected if:
	//   - the commits don't form a linear history.
	//   - the current head of from repo, once filtered, is not the immediate
	//     parent of those commits.
	//   - the modification contained is rejected by the filter.
	//   - the commits contains gpg signatures (can be turned off).
	CommitsFromSubRepo(ctx context.Context, in *CommitsFromSubRepoRequest, opts ...grpc.CallOption) (*CommitsFromSubRepoResponse, error)
	// CheckRepoSyncUpToDate checks if the head of current from repo, once
	// fitlered, is contained in the history of branch.
	CheckRepoSyncUpToDate(ctx context.Context, in *CheckRepoSyncUpToDateRequest, opts ...grpc.CallOption) (*CheckRepoSyncUpToDateResponse, error)
	// CheckCommitsFromSubRepo checks if the commits will be accepted into the
	// original repo.
	//
	// The commits will be rejected if:
	//   - the commits don't form a linear history.
	//   - the current head of from repo, once filtered, is not the immediate
	//     parent of those commits.
	//   - the modification contained is rejected by the filter.
	//   - the commits contains gpg signatures (can be turned off).
	CheckCommitsFromSubRepo(ctx context.Context, in *CheckCommitsFromSubRepoRequest, opts ...grpc.CallOption) (*CheckCommitsFromSubRepoResponse, error)
	// GetRepoSync obtain the sync relation by the id.
	GetRepoSync(ctx context.Context, in *GetRepoSyncRequest, opts ...grpc.CallOption) (*GetRepoSyncResponse, error)
}

type giTrimClient struct {
	cc grpc.ClientConnInterface
}

func NewGiTrimClient(cc grpc.ClientConnInterface) GiTrimClient {
	return &giTrimClient{cc}
}

func (c *giTrimClient) InitRepoSync(ctx context.Context, in *InitRepoSyncRequest, opts ...grpc.CallOption) (*InitRepoSyncResponse, error) {
	out := new(InitRepoSyncResponse)
	err := c.cc.Invoke(ctx, GiTrim_InitRepoSync_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *giTrimClient) SyncToSubRepo(ctx context.Context, in *SyncToSubRepoRequest, opts ...grpc.CallOption) (*SyncToSubRepoResponse, error) {
	out := new(SyncToSubRepoResponse)
	err := c.cc.Invoke(ctx, GiTrim_SyncToSubRepo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *giTrimClient) CommitsFromSubRepo(ctx context.Context, in *CommitsFromSubRepoRequest, opts ...grpc.CallOption) (*CommitsFromSubRepoResponse, error) {
	out := new(CommitsFromSubRepoResponse)
	err := c.cc.Invoke(ctx, GiTrim_CommitsFromSubRepo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *giTrimClient) CheckRepoSyncUpToDate(ctx context.Context, in *CheckRepoSyncUpToDateRequest, opts ...grpc.CallOption) (*CheckRepoSyncUpToDateResponse, error) {
	out := new(CheckRepoSyncUpToDateResponse)
	err := c.cc.Invoke(ctx, GiTrim_CheckRepoSyncUpToDate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *giTrimClient) CheckCommitsFromSubRepo(ctx context.Context, in *CheckCommitsFromSubRepoRequest, opts ...grpc.CallOption) (*CheckCommitsFromSubRepoResponse, error) {
	out := new(CheckCommitsFromSubRepoResponse)
	err := c.cc.Invoke(ctx, GiTrim_CheckCommitsFromSubRepo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *giTrimClient) GetRepoSync(ctx context.Context, in *GetRepoSyncRequest, opts ...grpc.CallOption) (*GetRepoSyncResponse, error) {
	out := new(GetRepoSyncResponse)
	err := c.cc.Invoke(ctx, GiTrim_GetRepoSync_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GiTrimServer is the server API for GiTrim service.
// All implementations must embed UnimplementedGiTrimServer
// for forward compatibility
type GiTrimServer interface {
	// InitRepoSync setup the sync-ing between two repos.
	//
	// The ID of the created repo will be sha256 hash of the following string
	//
	//	(from remote name)-(from owner name)-(from repo name)-(from branch)-(to
	//	remote name)-(to owner name)-(to repo name)-(to branch)
	//
	// The operation will also generate a secret for git webhooks, the secret is
	// generated from the id + 16 byte long salt.
	InitRepoSync(context.Context, *InitRepoSyncRequest) (*InitRepoSyncResponse, error)
	// SyncToSubRepo syncs from the original repo to the sub repo.
	// The request can be set to force.
	SyncToSubRepo(context.Context, *SyncToSubRepoRequest) (*SyncToSubRepoResponse, error)
	// CommitsFromSubRepo tries to sends a series of commits from a sub repo to
	// the original repo.
	//
	// The commits will be rejected if:
	//   - the commits don't form a linear history.
	//   - the current head of from repo, once filtered, is not the immediate
	//     parent of those commits.
	//   - the modification contained is rejected by the filter.
	//   - the commits contains gpg signatures (can be turned off).
	CommitsFromSubRepo(context.Context, *CommitsFromSubRepoRequest) (*CommitsFromSubRepoResponse, error)
	// CheckRepoSyncUpToDate checks if the head of current from repo, once
	// fitlered, is contained in the history of branch.
	CheckRepoSyncUpToDate(context.Context, *CheckRepoSyncUpToDateRequest) (*CheckRepoSyncUpToDateResponse, error)
	// CheckCommitsFromSubRepo checks if the commits will be accepted into the
	// original repo.
	//
	// The commits will be rejected if:
	//   - the commits don't form a linear history.
	//   - the current head of from repo, once filtered, is not the immediate
	//     parent of those commits.
	//   - the modification contained is rejected by the filter.
	//   - the commits contains gpg signatures (can be turned off).
	CheckCommitsFromSubRepo(context.Context, *CheckCommitsFromSubRepoRequest) (*CheckCommitsFromSubRepoResponse, error)
	// GetRepoSync obtain the sync relation by the id.
	GetRepoSync(context.Context, *GetRepoSyncRequest) (*GetRepoSyncResponse, error)
	mustEmbedUnimplementedGiTrimServer()
}

// UnimplementedGiTrimServer must be embedded to have forward compatible implementations.
type UnimplementedGiTrimServer struct {
}

func (UnimplementedGiTrimServer) InitRepoSync(context.Context, *InitRepoSyncRequest) (*InitRepoSyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitRepoSync not implemented")
}
func (UnimplementedGiTrimServer) SyncToSubRepo(context.Context, *SyncToSubRepoRequest) (*SyncToSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncToSubRepo not implemented")
}
func (UnimplementedGiTrimServer) CommitsFromSubRepo(context.Context, *CommitsFromSubRepoRequest) (*CommitsFromSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitsFromSubRepo not implemented")
}
func (UnimplementedGiTrimServer) CheckRepoSyncUpToDate(context.Context, *CheckRepoSyncUpToDateRequest) (*CheckRepoSyncUpToDateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckRepoSyncUpToDate not implemented")
}
func (UnimplementedGiTrimServer) CheckCommitsFromSubRepo(context.Context, *CheckCommitsFromSubRepoRequest) (*CheckCommitsFromSubRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckCommitsFromSubRepo not implemented")
}
func (UnimplementedGiTrimServer) GetRepoSync(context.Context, *GetRepoSyncRequest) (*GetRepoSyncResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRepoSync not implemented")
}
func (UnimplementedGiTrimServer) mustEmbedUnimplementedGiTrimServer() {}

// UnsafeGiTrimServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GiTrimServer will
// result in compilation errors.
type UnsafeGiTrimServer interface {
	mustEmbedUnimplementedGiTrimServer()
}

func RegisterGiTrimServer(s grpc.ServiceRegistrar, srv GiTrimServer) {
	s.RegisterService(&GiTrim_ServiceDesc, srv)
}

func _GiTrim_InitRepoSync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitRepoSyncRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GiTrimServer).InitRepoSync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GiTrim_InitRepoSync_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GiTrimServer).InitRepoSync(ctx, req.(*InitRepoSyncRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GiTrim_SyncToSubRepo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SyncToSubRepoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GiTrimServer).SyncToSubRepo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GiTrim_SyncToSubRepo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GiTrimServer).SyncToSubRepo(ctx, req.(*SyncToSubRepoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GiTrim_CommitsFromSubRepo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitsFromSubRepoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GiTrimServer).CommitsFromSubRepo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GiTrim_CommitsFromSubRepo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GiTrimServer).CommitsFromSubRepo(ctx, req.(*CommitsFromSubRepoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GiTrim_CheckRepoSyncUpToDate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckRepoSyncUpToDateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GiTrimServer).CheckRepoSyncUpToDate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GiTrim_CheckRepoSyncUpToDate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GiTrimServer).CheckRepoSyncUpToDate(ctx, req.(*CheckRepoSyncUpToDateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GiTrim_CheckCommitsFromSubRepo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckCommitsFromSubRepoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GiTrimServer).CheckCommitsFromSubRepo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GiTrim_CheckCommitsFromSubRepo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GiTrimServer).CheckCommitsFromSubRepo(ctx, req.(*CheckCommitsFromSubRepoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GiTrim_GetRepoSync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRepoSyncRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GiTrimServer).GetRepoSync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: GiTrim_GetRepoSync_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GiTrimServer).GetRepoSync(ctx, req.(*GetRepoSyncRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GiTrim_ServiceDesc is the grpc.ServiceDesc for GiTrim service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GiTrim_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gitrim.svc.GiTrim",
	HandlerType: (*GiTrimServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InitRepoSync",
			Handler:    _GiTrim_InitRepoSync_Handler,
		},
		{
			MethodName: "SyncToSubRepo",
			Handler:    _GiTrim_SyncToSubRepo_Handler,
		},
		{
			MethodName: "CommitsFromSubRepo",
			Handler:    _GiTrim_CommitsFromSubRepo_Handler,
		},
		{
			MethodName: "CheckRepoSyncUpToDate",
			Handler:    _GiTrim_CheckRepoSyncUpToDate_Handler,
		},
		{
			MethodName: "CheckCommitsFromSubRepo",
			Handler:    _GiTrim_CheckCommitsFromSubRepo_Handler,
		},
		{
			MethodName: "GetRepoSync",
			Handler:    _GiTrim_GetRepoSync_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "svc.proto",
}
