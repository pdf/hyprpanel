package panelplugin

import (
	"context"
	"math"
	"time"

	"github.com/hashicorp/go-plugin"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"google.golang.org/grpc"
)

// PanelGRPCClient panel plugin client implementation.
type PanelGRPCClient struct {
	broker *plugin.GRPCBroker
	client hyprpanelv1.PanelServiceClient
	server *grpc.Server
	ctx    context.Context
}

// Init implementation.
func (c *PanelGRPCClient) Init(h Host, id string, loglevel configv1.LogLevel, config *configv1.Panel, stylesheet []byte) error {
	host := &HostGRPCServer{Impl: h}
	serverFunc := func(opts []grpc.ServerOption) *grpc.Server {
		opts = append(opts, grpc.MaxRecvMsgSize(math.MaxInt32))
		opts = append(opts, grpc.MaxSendMsgSize(math.MaxInt32))
		c.server = grpc.NewServer(opts...)
		hyprpanelv1.RegisterHostServiceServer(c.server, host)
		return c.server
	}

	hostRef := c.broker.NextId()
	go c.broker.AcceptAndServe(hostRef, serverFunc)

	_, err := c.client.Init(context.Background(), &hyprpanelv1.PanelServiceInitRequest{
		Host:       hostRef,
		Id:         id,
		LogLevel:   loglevel,
		Config:     config,
		Stylesheet: stylesheet,
	})

	return err
}

// Notify implementation.
func (c *PanelGRPCClient) Notify(evt *eventv1.Event) {
	_, _ = c.client.Notify(context.Background(), &hyprpanelv1.PanelServiceNotifyRequest{Event: evt})
}

// Context implementation.
func (c *PanelGRPCClient) Context() context.Context {
	return c.ctx
}

// Close implementation.
func (c *PanelGRPCClient) Close() {
	defer c.server.Stop()
	c.client.Close(context.Background(), &hyprpanelv1.PanelServiceCloseRequest{})
}

// PanelGRPCServer panel plugin server implementation.
type PanelGRPCServer struct {
	hyprpanelv1.UnimplementedPanelServiceServer
	Impl Panel

	broker *plugin.GRPCBroker
	conn   *grpc.ClientConn
}

// Init implementation.
func (s *PanelGRPCServer) Init(_ context.Context, req *hyprpanelv1.PanelServiceInitRequest) (*hyprpanelv1.PanelServiceInitResponse, error) {
	var err error
	s.conn, err = s.broker.Dial(req.Host)
	if err != nil {
		return &hyprpanelv1.PanelServiceInitResponse{}, err
	}

	host := &HostGRPCClient{client: hyprpanelv1.NewHostServiceClient(s.conn)}
	if err := s.Impl.Init(host, req.Id, req.LogLevel, req.Config, req.Stylesheet); err != nil {
		return &hyprpanelv1.PanelServiceInitResponse{}, err
	}

	return &hyprpanelv1.PanelServiceInitResponse{}, nil
}

// Notify implementation.
func (s *PanelGRPCServer) Notify(_ context.Context, req *hyprpanelv1.PanelServiceNotifyRequest) (*hyprpanelv1.PanelServiceNotifyResponse, error) {
	s.Impl.Notify(req.Event)

	return &hyprpanelv1.PanelServiceNotifyResponse{}, nil
}

// Close implmenetation.
func (s *PanelGRPCServer) Close(_ context.Context, _ *hyprpanelv1.PanelServiceCloseRequest) (*hyprpanelv1.PanelServiceCloseResponse, error) {
	s.Impl.Close()
	if err := s.conn.Close(); err != nil {
		return &hyprpanelv1.PanelServiceCloseResponse{}, err
	}

	return &hyprpanelv1.PanelServiceCloseResponse{}, nil
}

// HostGRPCClient plugin host client implementation.
type HostGRPCClient struct {
	client hyprpanelv1.HostServiceClient
}

// Exec implmenetation.
func (c *HostGRPCClient) Exec(command string) error {
	_, err := c.client.Exec(context.Background(), &hyprpanelv1.HostServiceExecRequest{
		Command: command,
	})
	return err
}

// FindApplication implementation.
func (c *HostGRPCClient) FindApplication(query string) (*hyprpanelv1.AppInfo, error) {
	response, err := c.client.FindApplication(context.Background(), &hyprpanelv1.HostServiceFindApplicationRequest{
		Query: query,
	})
	if err != nil {
		return &hyprpanelv1.AppInfo{}, err
	}
	return response.AppInfo, nil
}

// SystrayActivate implementation.
func (c *HostGRPCClient) SystrayActivate(busName string, x, y int32) error {
	_, err := c.client.SystrayActivate(context.Background(), &hyprpanelv1.HostServiceSystrayActivateRequest{
		BusName: busName,
		X:       x,
		Y:       y,
	})
	return err
}

// SystraySecondaryActivate implmenetation.
func (c *HostGRPCClient) SystraySecondaryActivate(busName string, x, y int32) error {
	_, err := c.client.SystraySecondaryActivate(context.Background(), &hyprpanelv1.HostServiceSystraySecondaryActivateRequest{
		BusName: busName,
		X:       x,
		Y:       y,
	})
	return err
}

// SystrayScroll implementation.
func (c *HostGRPCClient) SystrayScroll(busName string, delta int32, orientation hyprpanelv1.SystrayScrollOrientation) error {
	_, err := c.client.SystrayScroll(context.Background(), &hyprpanelv1.HostServiceSystrayScrollRequest{
		BusName:     busName,
		Delta:       delta,
		Orientation: orientation,
	})
	return err
}

// SystrayMenuContextActivate implementation.
func (c *HostGRPCClient) SystrayMenuContextActivate(busName string, x, y int32) error {
	_, err := c.client.SystrayMenuContextActivate(context.Background(), &hyprpanelv1.HostServiceSystrayMenuContextActivateRequest{
		BusName: busName,
		X:       x,
		Y:       y,
	})
	return err
}

// SystrayMenuAboutToShow implementation.
func (c *HostGRPCClient) SystrayMenuAboutToShow(busName string, menuItemID string) error {
	_, err := c.client.SystrayMenuAboutToShow(context.Background(), &hyprpanelv1.HostServiceSystrayMenuAboutToShowRequest{
		BusName:    busName,
		MenuItemId: menuItemID,
	})
	return err
}

// SystrayMenuEvent implementation.
func (c *HostGRPCClient) SystrayMenuEvent(busName string, id int32, eventID hyprpanelv1.SystrayMenuEvent, _ any, timestamp time.Time) error {
	// TODO: Implement data field? Currently unused
	_, err := c.client.SystrayMenuEvent(context.Background(), &hyprpanelv1.HostServiceSystrayMenuEventRequest{
		BusName:   busName,
		Id:        id,
		EventId:   eventID,
		Data:      nil,
		Timestamp: uint32(timestamp.Unix()),
	})
	return err
}

// NotificationClosed implementation.
func (c *HostGRPCClient) NotificationClosed(id uint32, reason hyprpanelv1.NotificationClosedReason) error {
	_, err := c.client.NotificationClosed(context.Background(), &hyprpanelv1.HostServiceNotificationClosedRequest{
		Id:     id,
		Reason: reason,
	})
	return err
}

// NotificationAction implementation.
func (c *HostGRPCClient) NotificationAction(id uint32, actionKey string) error {
	_, err := c.client.NotificationAction(context.Background(), &hyprpanelv1.HostServiceNotificationActionRequest{
		Id:        id,
		ActionKey: actionKey,
	})
	return err
}

// AudioSinkVolumeAdjust implementation.
func (c *HostGRPCClient) AudioSinkVolumeAdjust(id string, direction eventv1.Direction) error {
	_, err := c.client.AudioSinkVolumeAdjust(context.Background(), &hyprpanelv1.HostServiceAudioSinkVolumeAdjustRequest{
		Id:        id,
		Direction: direction,
	})
	return err
}

// AudioSinkMuteToggle implementation.
func (c *HostGRPCClient) AudioSinkMuteToggle(id string) error {
	_, err := c.client.AudioSinkMuteToggle(context.Background(), &hyprpanelv1.HostServiceAudioSinkMuteToggleRequest{
		Id: id,
	})
	return err
}

// AudioSourceVolumeAdjust implementation.
func (c *HostGRPCClient) AudioSourceVolumeAdjust(id string, direction eventv1.Direction) error {
	_, err := c.client.AudioSourceVolumeAdjust(context.Background(), &hyprpanelv1.HostServiceAudioSourceVolumeAdjustRequest{
		Id:        id,
		Direction: direction,
	})
	return err
}

// AudioSourceMuteToggle implementation.
func (c *HostGRPCClient) AudioSourceMuteToggle(id string) error {
	_, err := c.client.AudioSourceMuteToggle(context.Background(), &hyprpanelv1.HostServiceAudioSourceMuteToggleRequest{
		Id: id,
	})
	return err
}

// BrightnessAdjust implementation.
func (c *HostGRPCClient) BrightnessAdjust(devName string, direction eventv1.Direction) error {
	_, err := c.client.BrightnessAdjust(context.Background(), &hyprpanelv1.HostServiceBrightnessAdjustRequest{
		DevName:   devName,
		Direction: direction,
	})
	return err
}

// CaptureFrame implementation.
func (c *HostGRPCClient) CaptureFrame(address uint64, width, height int32) (*hyprpanelv1.ImageNRGBA, error) {
	response, err := c.client.CaptureFrame(context.Background(), &hyprpanelv1.HostServiceCaptureFrameRequest{
		Address: address,
		Width:   width,
		Height:  height,
	})
	if err != nil {
		return &hyprpanelv1.ImageNRGBA{}, err
	}
	return response.Image, nil
}

// HostGRPCServer plugin host implementation.
type HostGRPCServer struct {
	hyprpanelv1.UnimplementedHostServiceServer
	Impl Host
}

// Exec implementation.
func (s *HostGRPCServer) Exec(_ context.Context, req *hyprpanelv1.HostServiceExecRequest) (*hyprpanelv1.HostServiceExecResponse, error) {
	err := s.Impl.Exec(req.Command)
	if err != nil {
		return &hyprpanelv1.HostServiceExecResponse{}, err
	}

	return &hyprpanelv1.HostServiceExecResponse{}, nil
}

// FindApplication implementation.
func (s *HostGRPCServer) FindApplication(_ context.Context, req *hyprpanelv1.HostServiceFindApplicationRequest) (*hyprpanelv1.HostServiceFindApplicationResponse, error) {
	appInfo, err := s.Impl.FindApplication(req.Query)
	if err != nil {
		return &hyprpanelv1.HostServiceFindApplicationResponse{}, err
	}

	return &hyprpanelv1.HostServiceFindApplicationResponse{
		AppInfo: appInfo,
	}, nil
}

// SystrayActivate implementation.
func (s *HostGRPCServer) SystrayActivate(_ context.Context, req *hyprpanelv1.HostServiceSystrayActivateRequest) (*hyprpanelv1.HostServiceSystrayActivateResponse, error) {
	err := s.Impl.SystrayActivate(req.BusName, req.X, req.Y)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayActivateResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayActivateResponse{}, nil
}

// SystraySecondaryActivate implementation.
func (s *HostGRPCServer) SystraySecondaryActivate(_ context.Context, req *hyprpanelv1.HostServiceSystraySecondaryActivateRequest) (*hyprpanelv1.HostServiceSystraySecondaryActivateResponse, error) {
	err := s.Impl.SystraySecondaryActivate(req.BusName, req.X, req.Y)
	if err != nil {
		return &hyprpanelv1.HostServiceSystraySecondaryActivateResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystraySecondaryActivateResponse{}, nil
}

// SystrayScroll implementation.
func (s *HostGRPCServer) SystrayScroll(_ context.Context, req *hyprpanelv1.HostServiceSystrayScrollRequest) (*hyprpanelv1.HostServiceSystrayScrollResponse, error) {
	err := s.Impl.SystrayScroll(req.BusName, req.Delta, req.Orientation)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayScrollResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayScrollResponse{}, nil
}

// SystrayMenuContextActivate implementation.
func (s *HostGRPCServer) SystrayMenuContextActivate(_ context.Context, req *hyprpanelv1.HostServiceSystrayMenuContextActivateRequest) (*hyprpanelv1.HostServiceSystrayMenuContextActivateResponse, error) {
	err := s.Impl.SystrayMenuContextActivate(req.BusName, req.X, req.Y)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayMenuContextActivateResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayMenuContextActivateResponse{}, nil
}

// SystrayMenuAboutToShow implementation.
func (s *HostGRPCServer) SystrayMenuAboutToShow(_ context.Context, req *hyprpanelv1.HostServiceSystrayMenuAboutToShowRequest) (*hyprpanelv1.HostServiceSystrayMenuAboutToShowResponse, error) {
	err := s.Impl.SystrayMenuAboutToShow(req.BusName, req.MenuItemId)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayMenuAboutToShowResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayMenuAboutToShowResponse{}, nil
}

// SystrayMenuEvent implementation.
func (s *HostGRPCServer) SystrayMenuEvent(_ context.Context, req *hyprpanelv1.HostServiceSystrayMenuEventRequest) (*hyprpanelv1.HostServiceSystrayMenuEventResponse, error) {
	timestamp := time.Unix(int64(req.Timestamp), 0)
	err := s.Impl.SystrayMenuEvent(req.BusName, req.Id, req.EventId, req.Data, timestamp)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayMenuEventResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayMenuEventResponse{}, nil
}

// NotificationClosed implementation.
func (s *HostGRPCServer) NotificationClosed(_ context.Context, req *hyprpanelv1.HostServiceNotificationClosedRequest) (*hyprpanelv1.HostServiceNotificationClosedResponse, error) {
	err := s.Impl.NotificationClosed(req.Id, req.Reason)
	if err != nil {
		return &hyprpanelv1.HostServiceNotificationClosedResponse{}, err
	}

	return &hyprpanelv1.HostServiceNotificationClosedResponse{}, nil
}

// NotificationAction implementation.
func (s *HostGRPCServer) NotificationAction(_ context.Context, req *hyprpanelv1.HostServiceNotificationActionRequest) (*hyprpanelv1.HostServiceNotificationActionResponse, error) {
	err := s.Impl.NotificationAction(req.Id, req.ActionKey)
	if err != nil {
		return &hyprpanelv1.HostServiceNotificationActionResponse{}, err
	}

	return &hyprpanelv1.HostServiceNotificationActionResponse{}, nil
}

// AudioSinkVolumeAdjust implementation.
func (s *HostGRPCServer) AudioSinkVolumeAdjust(_ context.Context, req *hyprpanelv1.HostServiceAudioSinkVolumeAdjustRequest) (*hyprpanelv1.HostServiceAudioSinkVolumeAdjustResponse, error) {
	err := s.Impl.AudioSinkVolumeAdjust(req.Id, req.Direction)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSinkVolumeAdjustResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSinkVolumeAdjustResponse{}, nil
}

// AudioSinkMuteToggle implementation.
func (s *HostGRPCServer) AudioSinkMuteToggle(_ context.Context, req *hyprpanelv1.HostServiceAudioSinkMuteToggleRequest) (*hyprpanelv1.HostServiceAudioSinkMuteToggleResponse, error) {
	err := s.Impl.AudioSinkMuteToggle(req.Id)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSinkMuteToggleResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSinkMuteToggleResponse{}, nil
}

// AudioSourceVolumeAdjust implementation.
func (s *HostGRPCServer) AudioSourceVolumeAdjust(_ context.Context, req *hyprpanelv1.HostServiceAudioSourceVolumeAdjustRequest) (*hyprpanelv1.HostServiceAudioSourceVolumeAdjustResponse, error) {
	err := s.Impl.AudioSourceVolumeAdjust(req.Id, req.Direction)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSourceVolumeAdjustResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSourceVolumeAdjustResponse{}, nil
}

// AudioSourceMuteToggle implmenetation.
func (s *HostGRPCServer) AudioSourceMuteToggle(_ context.Context, req *hyprpanelv1.HostServiceAudioSourceMuteToggleRequest) (*hyprpanelv1.HostServiceAudioSourceMuteToggleResponse, error) {
	err := s.Impl.AudioSourceMuteToggle(req.Id)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSourceMuteToggleResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSourceMuteToggleResponse{}, nil
}

// BrightnessAdjust implementation.
func (s *HostGRPCServer) BrightnessAdjust(_ context.Context, req *hyprpanelv1.HostServiceBrightnessAdjustRequest) (*hyprpanelv1.HostServiceBrightnessAdjustResponse, error) {
	err := s.Impl.BrightnessAdjust(req.DevName, req.Direction)
	if err != nil {
		return &hyprpanelv1.HostServiceBrightnessAdjustResponse{}, err
	}

	return &hyprpanelv1.HostServiceBrightnessAdjustResponse{}, nil
}

// CaptureFrame implementation.
func (s *HostGRPCServer) CaptureFrame(_ context.Context, req *hyprpanelv1.HostServiceCaptureFrameRequest) (*hyprpanelv1.HostServiceCaptureFrameResponse, error) {
	img, err := s.Impl.CaptureFrame(req.Address, req.Width, req.Height)
	if err != nil {
		return &hyprpanelv1.HostServiceCaptureFrameResponse{}, err
	}

	return &hyprpanelv1.HostServiceCaptureFrameResponse{
		Image: img,
	}, nil
}
