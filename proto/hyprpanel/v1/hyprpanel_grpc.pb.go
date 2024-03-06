// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: hyprpanel/v1/hyprpanel.proto

package hyprpanelv1

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
	PanelService_Init_FullMethodName   = "/hyprpanel.v1.PanelService/Init"
	PanelService_Notify_FullMethodName = "/hyprpanel.v1.PanelService/Notify"
	PanelService_Close_FullMethodName  = "/hyprpanel.v1.PanelService/Close"
)

// PanelServiceClient is the client API for PanelService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PanelServiceClient interface {
	Init(ctx context.Context, in *PanelServiceInitRequest, opts ...grpc.CallOption) (*PanelServiceInitResponse, error)
	Notify(ctx context.Context, in *PanelServiceNotifyRequest, opts ...grpc.CallOption) (*PanelServiceNotifyResponse, error)
	Close(ctx context.Context, in *PanelServiceCloseRequest, opts ...grpc.CallOption) (*PanelServiceCloseResponse, error)
}

type panelServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPanelServiceClient(cc grpc.ClientConnInterface) PanelServiceClient {
	return &panelServiceClient{cc}
}

func (c *panelServiceClient) Init(ctx context.Context, in *PanelServiceInitRequest, opts ...grpc.CallOption) (*PanelServiceInitResponse, error) {
	out := new(PanelServiceInitResponse)
	err := c.cc.Invoke(ctx, PanelService_Init_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *panelServiceClient) Notify(ctx context.Context, in *PanelServiceNotifyRequest, opts ...grpc.CallOption) (*PanelServiceNotifyResponse, error) {
	out := new(PanelServiceNotifyResponse)
	err := c.cc.Invoke(ctx, PanelService_Notify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *panelServiceClient) Close(ctx context.Context, in *PanelServiceCloseRequest, opts ...grpc.CallOption) (*PanelServiceCloseResponse, error) {
	out := new(PanelServiceCloseResponse)
	err := c.cc.Invoke(ctx, PanelService_Close_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PanelServiceServer is the server API for PanelService service.
// All implementations must embed UnimplementedPanelServiceServer
// for forward compatibility
type PanelServiceServer interface {
	Init(context.Context, *PanelServiceInitRequest) (*PanelServiceInitResponse, error)
	Notify(context.Context, *PanelServiceNotifyRequest) (*PanelServiceNotifyResponse, error)
	Close(context.Context, *PanelServiceCloseRequest) (*PanelServiceCloseResponse, error)
	mustEmbedUnimplementedPanelServiceServer()
}

// UnimplementedPanelServiceServer must be embedded to have forward compatible implementations.
type UnimplementedPanelServiceServer struct {
}

func (UnimplementedPanelServiceServer) Init(context.Context, *PanelServiceInitRequest) (*PanelServiceInitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Init not implemented")
}
func (UnimplementedPanelServiceServer) Notify(context.Context, *PanelServiceNotifyRequest) (*PanelServiceNotifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Notify not implemented")
}
func (UnimplementedPanelServiceServer) Close(context.Context, *PanelServiceCloseRequest) (*PanelServiceCloseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Close not implemented")
}
func (UnimplementedPanelServiceServer) mustEmbedUnimplementedPanelServiceServer() {}

// UnsafePanelServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PanelServiceServer will
// result in compilation errors.
type UnsafePanelServiceServer interface {
	mustEmbedUnimplementedPanelServiceServer()
}

func RegisterPanelServiceServer(s grpc.ServiceRegistrar, srv PanelServiceServer) {
	s.RegisterService(&PanelService_ServiceDesc, srv)
}

func _PanelService_Init_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PanelServiceInitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PanelServiceServer).Init(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PanelService_Init_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PanelServiceServer).Init(ctx, req.(*PanelServiceInitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PanelService_Notify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PanelServiceNotifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PanelServiceServer).Notify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PanelService_Notify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PanelServiceServer).Notify(ctx, req.(*PanelServiceNotifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PanelService_Close_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PanelServiceCloseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PanelServiceServer).Close(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PanelService_Close_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PanelServiceServer).Close(ctx, req.(*PanelServiceCloseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PanelService_ServiceDesc is the grpc.ServiceDesc for PanelService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PanelService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "hyprpanel.v1.PanelService",
	HandlerType: (*PanelServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Init",
			Handler:    _PanelService_Init_Handler,
		},
		{
			MethodName: "Notify",
			Handler:    _PanelService_Notify_Handler,
		},
		{
			MethodName: "Close",
			Handler:    _PanelService_Close_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hyprpanel/v1/hyprpanel.proto",
}

const (
	HostService_Exec_FullMethodName                       = "/hyprpanel.v1.HostService/Exec"
	HostService_FindApplication_FullMethodName            = "/hyprpanel.v1.HostService/FindApplication"
	HostService_SystrayActivate_FullMethodName            = "/hyprpanel.v1.HostService/SystrayActivate"
	HostService_SystraySecondaryActivate_FullMethodName   = "/hyprpanel.v1.HostService/SystraySecondaryActivate"
	HostService_SystrayScroll_FullMethodName              = "/hyprpanel.v1.HostService/SystrayScroll"
	HostService_SystrayMenuContextActivate_FullMethodName = "/hyprpanel.v1.HostService/SystrayMenuContextActivate"
	HostService_SystrayMenuAboutToShow_FullMethodName     = "/hyprpanel.v1.HostService/SystrayMenuAboutToShow"
	HostService_SystrayMenuEvent_FullMethodName           = "/hyprpanel.v1.HostService/SystrayMenuEvent"
	HostService_NotificationClosed_FullMethodName         = "/hyprpanel.v1.HostService/NotificationClosed"
	HostService_NotificationAction_FullMethodName         = "/hyprpanel.v1.HostService/NotificationAction"
	HostService_AudioSinkVolumeAdjust_FullMethodName      = "/hyprpanel.v1.HostService/AudioSinkVolumeAdjust"
	HostService_AudioSinkMuteToggle_FullMethodName        = "/hyprpanel.v1.HostService/AudioSinkMuteToggle"
	HostService_AudioSourceVolumeAdjust_FullMethodName    = "/hyprpanel.v1.HostService/AudioSourceVolumeAdjust"
	HostService_AudioSourceMuteToggle_FullMethodName      = "/hyprpanel.v1.HostService/AudioSourceMuteToggle"
	HostService_BrightnessAdjust_FullMethodName           = "/hyprpanel.v1.HostService/BrightnessAdjust"
)

// HostServiceClient is the client API for HostService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HostServiceClient interface {
	Exec(ctx context.Context, in *HostServiceExecRequest, opts ...grpc.CallOption) (*HostServiceExecResponse, error)
	FindApplication(ctx context.Context, in *HostServiceFindApplicationRequest, opts ...grpc.CallOption) (*HostServiceFindApplicationResponse, error)
	SystrayActivate(ctx context.Context, in *HostServiceSystrayActivateRequest, opts ...grpc.CallOption) (*HostServiceSystrayActivateResponse, error)
	SystraySecondaryActivate(ctx context.Context, in *HostServiceSystraySecondaryActivateRequest, opts ...grpc.CallOption) (*HostServiceSystraySecondaryActivateResponse, error)
	SystrayScroll(ctx context.Context, in *HostServiceSystrayScrollRequest, opts ...grpc.CallOption) (*HostServiceSystrayScrollResponse, error)
	SystrayMenuContextActivate(ctx context.Context, in *HostServiceSystrayMenuContextActivateRequest, opts ...grpc.CallOption) (*HostServiceSystrayMenuContextActivateResponse, error)
	SystrayMenuAboutToShow(ctx context.Context, in *HostServiceSystrayMenuAboutToShowRequest, opts ...grpc.CallOption) (*HostServiceSystrayMenuAboutToShowResponse, error)
	SystrayMenuEvent(ctx context.Context, in *HostServiceSystrayMenuEventRequest, opts ...grpc.CallOption) (*HostServiceSystrayMenuEventResponse, error)
	NotificationClosed(ctx context.Context, in *HostServiceNotificationClosedRequest, opts ...grpc.CallOption) (*HostServiceNotificationClosedResponse, error)
	NotificationAction(ctx context.Context, in *HostServiceNotificationActionRequest, opts ...grpc.CallOption) (*HostServiceNotificationActionResponse, error)
	AudioSinkVolumeAdjust(ctx context.Context, in *HostServiceAudioSinkVolumeAdjustRequest, opts ...grpc.CallOption) (*HostServiceAudioSinkVolumeAdjustResponse, error)
	AudioSinkMuteToggle(ctx context.Context, in *HostServiceAudioSinkMuteToggleRequest, opts ...grpc.CallOption) (*HostServiceAudioSinkMuteToggleResponse, error)
	AudioSourceVolumeAdjust(ctx context.Context, in *HostServiceAudioSourceVolumeAdjustRequest, opts ...grpc.CallOption) (*HostServiceAudioSourceVolumeAdjustResponse, error)
	AudioSourceMuteToggle(ctx context.Context, in *HostServiceAudioSourceMuteToggleRequest, opts ...grpc.CallOption) (*HostServiceAudioSourceMuteToggleResponse, error)
	BrightnessAdjust(ctx context.Context, in *HostServiceBrightnessAdjustRequest, opts ...grpc.CallOption) (*HostServiceBrightnessAdjustResponse, error)
}

type hostServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHostServiceClient(cc grpc.ClientConnInterface) HostServiceClient {
	return &hostServiceClient{cc}
}

func (c *hostServiceClient) Exec(ctx context.Context, in *HostServiceExecRequest, opts ...grpc.CallOption) (*HostServiceExecResponse, error) {
	out := new(HostServiceExecResponse)
	err := c.cc.Invoke(ctx, HostService_Exec_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) FindApplication(ctx context.Context, in *HostServiceFindApplicationRequest, opts ...grpc.CallOption) (*HostServiceFindApplicationResponse, error) {
	out := new(HostServiceFindApplicationResponse)
	err := c.cc.Invoke(ctx, HostService_FindApplication_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) SystrayActivate(ctx context.Context, in *HostServiceSystrayActivateRequest, opts ...grpc.CallOption) (*HostServiceSystrayActivateResponse, error) {
	out := new(HostServiceSystrayActivateResponse)
	err := c.cc.Invoke(ctx, HostService_SystrayActivate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) SystraySecondaryActivate(ctx context.Context, in *HostServiceSystraySecondaryActivateRequest, opts ...grpc.CallOption) (*HostServiceSystraySecondaryActivateResponse, error) {
	out := new(HostServiceSystraySecondaryActivateResponse)
	err := c.cc.Invoke(ctx, HostService_SystraySecondaryActivate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) SystrayScroll(ctx context.Context, in *HostServiceSystrayScrollRequest, opts ...grpc.CallOption) (*HostServiceSystrayScrollResponse, error) {
	out := new(HostServiceSystrayScrollResponse)
	err := c.cc.Invoke(ctx, HostService_SystrayScroll_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) SystrayMenuContextActivate(ctx context.Context, in *HostServiceSystrayMenuContextActivateRequest, opts ...grpc.CallOption) (*HostServiceSystrayMenuContextActivateResponse, error) {
	out := new(HostServiceSystrayMenuContextActivateResponse)
	err := c.cc.Invoke(ctx, HostService_SystrayMenuContextActivate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) SystrayMenuAboutToShow(ctx context.Context, in *HostServiceSystrayMenuAboutToShowRequest, opts ...grpc.CallOption) (*HostServiceSystrayMenuAboutToShowResponse, error) {
	out := new(HostServiceSystrayMenuAboutToShowResponse)
	err := c.cc.Invoke(ctx, HostService_SystrayMenuAboutToShow_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) SystrayMenuEvent(ctx context.Context, in *HostServiceSystrayMenuEventRequest, opts ...grpc.CallOption) (*HostServiceSystrayMenuEventResponse, error) {
	out := new(HostServiceSystrayMenuEventResponse)
	err := c.cc.Invoke(ctx, HostService_SystrayMenuEvent_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) NotificationClosed(ctx context.Context, in *HostServiceNotificationClosedRequest, opts ...grpc.CallOption) (*HostServiceNotificationClosedResponse, error) {
	out := new(HostServiceNotificationClosedResponse)
	err := c.cc.Invoke(ctx, HostService_NotificationClosed_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) NotificationAction(ctx context.Context, in *HostServiceNotificationActionRequest, opts ...grpc.CallOption) (*HostServiceNotificationActionResponse, error) {
	out := new(HostServiceNotificationActionResponse)
	err := c.cc.Invoke(ctx, HostService_NotificationAction_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) AudioSinkVolumeAdjust(ctx context.Context, in *HostServiceAudioSinkVolumeAdjustRequest, opts ...grpc.CallOption) (*HostServiceAudioSinkVolumeAdjustResponse, error) {
	out := new(HostServiceAudioSinkVolumeAdjustResponse)
	err := c.cc.Invoke(ctx, HostService_AudioSinkVolumeAdjust_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) AudioSinkMuteToggle(ctx context.Context, in *HostServiceAudioSinkMuteToggleRequest, opts ...grpc.CallOption) (*HostServiceAudioSinkMuteToggleResponse, error) {
	out := new(HostServiceAudioSinkMuteToggleResponse)
	err := c.cc.Invoke(ctx, HostService_AudioSinkMuteToggle_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) AudioSourceVolumeAdjust(ctx context.Context, in *HostServiceAudioSourceVolumeAdjustRequest, opts ...grpc.CallOption) (*HostServiceAudioSourceVolumeAdjustResponse, error) {
	out := new(HostServiceAudioSourceVolumeAdjustResponse)
	err := c.cc.Invoke(ctx, HostService_AudioSourceVolumeAdjust_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) AudioSourceMuteToggle(ctx context.Context, in *HostServiceAudioSourceMuteToggleRequest, opts ...grpc.CallOption) (*HostServiceAudioSourceMuteToggleResponse, error) {
	out := new(HostServiceAudioSourceMuteToggleResponse)
	err := c.cc.Invoke(ctx, HostService_AudioSourceMuteToggle_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hostServiceClient) BrightnessAdjust(ctx context.Context, in *HostServiceBrightnessAdjustRequest, opts ...grpc.CallOption) (*HostServiceBrightnessAdjustResponse, error) {
	out := new(HostServiceBrightnessAdjustResponse)
	err := c.cc.Invoke(ctx, HostService_BrightnessAdjust_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HostServiceServer is the server API for HostService service.
// All implementations must embed UnimplementedHostServiceServer
// for forward compatibility
type HostServiceServer interface {
	Exec(context.Context, *HostServiceExecRequest) (*HostServiceExecResponse, error)
	FindApplication(context.Context, *HostServiceFindApplicationRequest) (*HostServiceFindApplicationResponse, error)
	SystrayActivate(context.Context, *HostServiceSystrayActivateRequest) (*HostServiceSystrayActivateResponse, error)
	SystraySecondaryActivate(context.Context, *HostServiceSystraySecondaryActivateRequest) (*HostServiceSystraySecondaryActivateResponse, error)
	SystrayScroll(context.Context, *HostServiceSystrayScrollRequest) (*HostServiceSystrayScrollResponse, error)
	SystrayMenuContextActivate(context.Context, *HostServiceSystrayMenuContextActivateRequest) (*HostServiceSystrayMenuContextActivateResponse, error)
	SystrayMenuAboutToShow(context.Context, *HostServiceSystrayMenuAboutToShowRequest) (*HostServiceSystrayMenuAboutToShowResponse, error)
	SystrayMenuEvent(context.Context, *HostServiceSystrayMenuEventRequest) (*HostServiceSystrayMenuEventResponse, error)
	NotificationClosed(context.Context, *HostServiceNotificationClosedRequest) (*HostServiceNotificationClosedResponse, error)
	NotificationAction(context.Context, *HostServiceNotificationActionRequest) (*HostServiceNotificationActionResponse, error)
	AudioSinkVolumeAdjust(context.Context, *HostServiceAudioSinkVolumeAdjustRequest) (*HostServiceAudioSinkVolumeAdjustResponse, error)
	AudioSinkMuteToggle(context.Context, *HostServiceAudioSinkMuteToggleRequest) (*HostServiceAudioSinkMuteToggleResponse, error)
	AudioSourceVolumeAdjust(context.Context, *HostServiceAudioSourceVolumeAdjustRequest) (*HostServiceAudioSourceVolumeAdjustResponse, error)
	AudioSourceMuteToggle(context.Context, *HostServiceAudioSourceMuteToggleRequest) (*HostServiceAudioSourceMuteToggleResponse, error)
	BrightnessAdjust(context.Context, *HostServiceBrightnessAdjustRequest) (*HostServiceBrightnessAdjustResponse, error)
	mustEmbedUnimplementedHostServiceServer()
}

// UnimplementedHostServiceServer must be embedded to have forward compatible implementations.
type UnimplementedHostServiceServer struct {
}

func (UnimplementedHostServiceServer) Exec(context.Context, *HostServiceExecRequest) (*HostServiceExecResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Exec not implemented")
}
func (UnimplementedHostServiceServer) FindApplication(context.Context, *HostServiceFindApplicationRequest) (*HostServiceFindApplicationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindApplication not implemented")
}
func (UnimplementedHostServiceServer) SystrayActivate(context.Context, *HostServiceSystrayActivateRequest) (*HostServiceSystrayActivateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SystrayActivate not implemented")
}
func (UnimplementedHostServiceServer) SystraySecondaryActivate(context.Context, *HostServiceSystraySecondaryActivateRequest) (*HostServiceSystraySecondaryActivateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SystraySecondaryActivate not implemented")
}
func (UnimplementedHostServiceServer) SystrayScroll(context.Context, *HostServiceSystrayScrollRequest) (*HostServiceSystrayScrollResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SystrayScroll not implemented")
}
func (UnimplementedHostServiceServer) SystrayMenuContextActivate(context.Context, *HostServiceSystrayMenuContextActivateRequest) (*HostServiceSystrayMenuContextActivateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SystrayMenuContextActivate not implemented")
}
func (UnimplementedHostServiceServer) SystrayMenuAboutToShow(context.Context, *HostServiceSystrayMenuAboutToShowRequest) (*HostServiceSystrayMenuAboutToShowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SystrayMenuAboutToShow not implemented")
}
func (UnimplementedHostServiceServer) SystrayMenuEvent(context.Context, *HostServiceSystrayMenuEventRequest) (*HostServiceSystrayMenuEventResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SystrayMenuEvent not implemented")
}
func (UnimplementedHostServiceServer) NotificationClosed(context.Context, *HostServiceNotificationClosedRequest) (*HostServiceNotificationClosedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotificationClosed not implemented")
}
func (UnimplementedHostServiceServer) NotificationAction(context.Context, *HostServiceNotificationActionRequest) (*HostServiceNotificationActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotificationAction not implemented")
}
func (UnimplementedHostServiceServer) AudioSinkVolumeAdjust(context.Context, *HostServiceAudioSinkVolumeAdjustRequest) (*HostServiceAudioSinkVolumeAdjustResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AudioSinkVolumeAdjust not implemented")
}
func (UnimplementedHostServiceServer) AudioSinkMuteToggle(context.Context, *HostServiceAudioSinkMuteToggleRequest) (*HostServiceAudioSinkMuteToggleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AudioSinkMuteToggle not implemented")
}
func (UnimplementedHostServiceServer) AudioSourceVolumeAdjust(context.Context, *HostServiceAudioSourceVolumeAdjustRequest) (*HostServiceAudioSourceVolumeAdjustResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AudioSourceVolumeAdjust not implemented")
}
func (UnimplementedHostServiceServer) AudioSourceMuteToggle(context.Context, *HostServiceAudioSourceMuteToggleRequest) (*HostServiceAudioSourceMuteToggleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AudioSourceMuteToggle not implemented")
}
func (UnimplementedHostServiceServer) BrightnessAdjust(context.Context, *HostServiceBrightnessAdjustRequest) (*HostServiceBrightnessAdjustResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BrightnessAdjust not implemented")
}
func (UnimplementedHostServiceServer) mustEmbedUnimplementedHostServiceServer() {}

// UnsafeHostServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HostServiceServer will
// result in compilation errors.
type UnsafeHostServiceServer interface {
	mustEmbedUnimplementedHostServiceServer()
}

func RegisterHostServiceServer(s grpc.ServiceRegistrar, srv HostServiceServer) {
	s.RegisterService(&HostService_ServiceDesc, srv)
}

func _HostService_Exec_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceExecRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).Exec(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_Exec_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).Exec(ctx, req.(*HostServiceExecRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_FindApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceFindApplicationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).FindApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_FindApplication_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).FindApplication(ctx, req.(*HostServiceFindApplicationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_SystrayActivate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceSystrayActivateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).SystrayActivate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_SystrayActivate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).SystrayActivate(ctx, req.(*HostServiceSystrayActivateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_SystraySecondaryActivate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceSystraySecondaryActivateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).SystraySecondaryActivate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_SystraySecondaryActivate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).SystraySecondaryActivate(ctx, req.(*HostServiceSystraySecondaryActivateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_SystrayScroll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceSystrayScrollRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).SystrayScroll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_SystrayScroll_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).SystrayScroll(ctx, req.(*HostServiceSystrayScrollRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_SystrayMenuContextActivate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceSystrayMenuContextActivateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).SystrayMenuContextActivate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_SystrayMenuContextActivate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).SystrayMenuContextActivate(ctx, req.(*HostServiceSystrayMenuContextActivateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_SystrayMenuAboutToShow_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceSystrayMenuAboutToShowRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).SystrayMenuAboutToShow(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_SystrayMenuAboutToShow_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).SystrayMenuAboutToShow(ctx, req.(*HostServiceSystrayMenuAboutToShowRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_SystrayMenuEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceSystrayMenuEventRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).SystrayMenuEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_SystrayMenuEvent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).SystrayMenuEvent(ctx, req.(*HostServiceSystrayMenuEventRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_NotificationClosed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceNotificationClosedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).NotificationClosed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_NotificationClosed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).NotificationClosed(ctx, req.(*HostServiceNotificationClosedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_NotificationAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceNotificationActionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).NotificationAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_NotificationAction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).NotificationAction(ctx, req.(*HostServiceNotificationActionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_AudioSinkVolumeAdjust_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceAudioSinkVolumeAdjustRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).AudioSinkVolumeAdjust(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_AudioSinkVolumeAdjust_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).AudioSinkVolumeAdjust(ctx, req.(*HostServiceAudioSinkVolumeAdjustRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_AudioSinkMuteToggle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceAudioSinkMuteToggleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).AudioSinkMuteToggle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_AudioSinkMuteToggle_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).AudioSinkMuteToggle(ctx, req.(*HostServiceAudioSinkMuteToggleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_AudioSourceVolumeAdjust_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceAudioSourceVolumeAdjustRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).AudioSourceVolumeAdjust(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_AudioSourceVolumeAdjust_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).AudioSourceVolumeAdjust(ctx, req.(*HostServiceAudioSourceVolumeAdjustRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_AudioSourceMuteToggle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceAudioSourceMuteToggleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).AudioSourceMuteToggle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_AudioSourceMuteToggle_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).AudioSourceMuteToggle(ctx, req.(*HostServiceAudioSourceMuteToggleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HostService_BrightnessAdjust_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HostServiceBrightnessAdjustRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HostServiceServer).BrightnessAdjust(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: HostService_BrightnessAdjust_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HostServiceServer).BrightnessAdjust(ctx, req.(*HostServiceBrightnessAdjustRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// HostService_ServiceDesc is the grpc.ServiceDesc for HostService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HostService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "hyprpanel.v1.HostService",
	HandlerType: (*HostServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Exec",
			Handler:    _HostService_Exec_Handler,
		},
		{
			MethodName: "FindApplication",
			Handler:    _HostService_FindApplication_Handler,
		},
		{
			MethodName: "SystrayActivate",
			Handler:    _HostService_SystrayActivate_Handler,
		},
		{
			MethodName: "SystraySecondaryActivate",
			Handler:    _HostService_SystraySecondaryActivate_Handler,
		},
		{
			MethodName: "SystrayScroll",
			Handler:    _HostService_SystrayScroll_Handler,
		},
		{
			MethodName: "SystrayMenuContextActivate",
			Handler:    _HostService_SystrayMenuContextActivate_Handler,
		},
		{
			MethodName: "SystrayMenuAboutToShow",
			Handler:    _HostService_SystrayMenuAboutToShow_Handler,
		},
		{
			MethodName: "SystrayMenuEvent",
			Handler:    _HostService_SystrayMenuEvent_Handler,
		},
		{
			MethodName: "NotificationClosed",
			Handler:    _HostService_NotificationClosed_Handler,
		},
		{
			MethodName: "NotificationAction",
			Handler:    _HostService_NotificationAction_Handler,
		},
		{
			MethodName: "AudioSinkVolumeAdjust",
			Handler:    _HostService_AudioSinkVolumeAdjust_Handler,
		},
		{
			MethodName: "AudioSinkMuteToggle",
			Handler:    _HostService_AudioSinkMuteToggle_Handler,
		},
		{
			MethodName: "AudioSourceVolumeAdjust",
			Handler:    _HostService_AudioSourceVolumeAdjust_Handler,
		},
		{
			MethodName: "AudioSourceMuteToggle",
			Handler:    _HostService_AudioSourceMuteToggle_Handler,
		},
		{
			MethodName: "BrightnessAdjust",
			Handler:    _HostService_BrightnessAdjust_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hyprpanel/v1/hyprpanel.proto",
}