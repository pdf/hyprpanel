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

type PanelGRPCClient struct {
	broker *plugin.GRPCBroker
	client hyprpanelv1.PanelServiceClient
	server *grpc.Server
	ctx    context.Context
}

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

	c.client.Init(context.Background(), &hyprpanelv1.PanelServiceInitRequest{
		Host:       hostRef,
		Id:         id,
		LogLevel:   loglevel,
		Config:     config,
		Stylesheet: stylesheet,
	})

	return nil
}

func (c *PanelGRPCClient) Notify(evt *eventv1.Event) {
	c.client.Notify(context.Background(), &hyprpanelv1.PanelServiceNotifyRequest{Event: evt})
}

func (c *PanelGRPCClient) Context() context.Context {
	return c.ctx
}

func (c *PanelGRPCClient) Close() {
	defer c.server.Stop()
	c.client.Close(context.Background(), &hyprpanelv1.PanelServiceCloseRequest{})
}

type PanelGRPCServer struct {
	hyprpanelv1.UnimplementedPanelServiceServer
	Impl Panel

	broker *plugin.GRPCBroker
	conn   *grpc.ClientConn
}

func (s *PanelGRPCServer) Init(ctx context.Context, req *hyprpanelv1.PanelServiceInitRequest) (*hyprpanelv1.PanelServiceInitResponse, error) {
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

func (s *PanelGRPCServer) Notify(ctx context.Context, req *hyprpanelv1.PanelServiceNotifyRequest) (*hyprpanelv1.PanelServiceNotifyResponse, error) {
	s.Impl.Notify(req.Event)

	return &hyprpanelv1.PanelServiceNotifyResponse{}, nil
}

func (s *PanelGRPCServer) Close(ctx context.Context, req *hyprpanelv1.PanelServiceCloseRequest) (*hyprpanelv1.PanelServiceCloseResponse, error) {
	s.Impl.Close()
	if err := s.conn.Close(); err != nil {
		return &hyprpanelv1.PanelServiceCloseResponse{}, err
	}

	return &hyprpanelv1.PanelServiceCloseResponse{}, nil
}

type HostGRPCClient struct {
	client hyprpanelv1.HostServiceClient
}

func (c *HostGRPCClient) Exec(command string) error {
	_, err := c.client.Exec(context.Background(), &hyprpanelv1.HostServiceExecRequest{
		Command: command,
	})
	return err
}

func (c *HostGRPCClient) FindApplication(query string) (*hyprpanelv1.AppInfo, error) {
	response, err := c.client.FindApplication(context.Background(), &hyprpanelv1.HostServiceFindApplicationRequest{
		Query: query,
	})
	if err != nil {
		return &hyprpanelv1.AppInfo{}, err
	}
	return response.AppInfo, nil
}

func (c *HostGRPCClient) SystrayActivate(busName string, x, y int32) error {
	_, err := c.client.SystrayActivate(context.Background(), &hyprpanelv1.HostServiceSystrayActivateRequest{
		BusName: busName,
		X:       x,
		Y:       y,
	})
	return err
}

func (c *HostGRPCClient) SystraySecondaryActivate(busName string, x, y int32) error {
	_, err := c.client.SystraySecondaryActivate(context.Background(), &hyprpanelv1.HostServiceSystraySecondaryActivateRequest{
		BusName: busName,
		X:       x,
		Y:       y,
	})
	return err
}

func (c *HostGRPCClient) SystrayScroll(busName string, delta int32, orientation hyprpanelv1.SystrayScrollOrientation) error {
	_, err := c.client.SystrayScroll(context.Background(), &hyprpanelv1.HostServiceSystrayScrollRequest{
		BusName:     busName,
		Delta:       delta,
		Orientation: orientation,
	})
	return err
}

func (c *HostGRPCClient) SystrayMenuContextActivate(busName string, x, y int32) error {
	_, err := c.client.SystrayMenuContextActivate(context.Background(), &hyprpanelv1.HostServiceSystrayMenuContextActivateRequest{
		BusName: busName,
		X:       x,
		Y:       y,
	})
	return err
}

func (c *HostGRPCClient) SystrayMenuAboutToShow(busName string, menuItemID string) error {
	_, err := c.client.SystrayMenuAboutToShow(context.Background(), &hyprpanelv1.HostServiceSystrayMenuAboutToShowRequest{
		BusName:    busName,
		MenuItemId: menuItemID,
	})
	return err
}

func (c *HostGRPCClient) SystrayMenuEvent(busName string, id int32, eventId hyprpanelv1.SystrayMenuEvent, data any, timestamp time.Time) error {
	// TODO: Implement data field? Currently unused
	_, err := c.client.SystrayMenuEvent(context.Background(), &hyprpanelv1.HostServiceSystrayMenuEventRequest{
		BusName:   busName,
		Id:        id,
		EventId:   eventId,
		Data:      nil,
		Timestamp: uint32(timestamp.Unix()),
	})
	return err
}

func (c *HostGRPCClient) NotificationClosed(id uint32, reason hyprpanelv1.NotificationClosedReason) error {
	_, err := c.client.NotificationClosed(context.Background(), &hyprpanelv1.HostServiceNotificationClosedRequest{
		Id:     id,
		Reason: reason,
	})
	return err
}

func (c *HostGRPCClient) NotificationAction(id uint32, actionKey string) error {
	_, err := c.client.NotificationAction(context.Background(), &hyprpanelv1.HostServiceNotificationActionRequest{
		Id:        id,
		ActionKey: actionKey,
	})
	return err
}

func (c *HostGRPCClient) AudioSinkVolumeAdjust(id string, direction eventv1.Direction) error {
	_, err := c.client.AudioSinkVolumeAdjust(context.Background(), &hyprpanelv1.HostServiceAudioSinkVolumeAdjustRequest{
		Id:        id,
		Direction: direction,
	})
	return err
}

func (c *HostGRPCClient) AudioSinkMuteToggle(id string) error {
	_, err := c.client.AudioSinkMuteToggle(context.Background(), &hyprpanelv1.HostServiceAudioSinkMuteToggleRequest{
		Id: id,
	})
	return err
}

func (c *HostGRPCClient) AudioSourceVolumeAdjust(id string, direction eventv1.Direction) error {
	_, err := c.client.AudioSourceVolumeAdjust(context.Background(), &hyprpanelv1.HostServiceAudioSourceVolumeAdjustRequest{
		Id:        id,
		Direction: direction,
	})
	return err
}

func (c *HostGRPCClient) AudioSourceMuteToggle(id string) error {
	_, err := c.client.AudioSourceMuteToggle(context.Background(), &hyprpanelv1.HostServiceAudioSourceMuteToggleRequest{
		Id: id,
	})
	return err
}

func (c *HostGRPCClient) BrightnessAdjust(devName string, direction eventv1.Direction) error {
	_, err := c.client.BrightnessAdjust(context.Background(), &hyprpanelv1.HostServiceBrightnessAdjustRequest{
		DevName:   devName,
		Direction: direction,
	})
	return err
}

type HostGRPCServer struct {
	hyprpanelv1.UnimplementedHostServiceServer
	Impl Host
}

func (s *HostGRPCServer) Exec(ctx context.Context, req *hyprpanelv1.HostServiceExecRequest) (*hyprpanelv1.HostServiceExecResponse, error) {
	err := s.Impl.Exec(req.Command)
	if err != nil {
		return &hyprpanelv1.HostServiceExecResponse{}, err
	}

	return &hyprpanelv1.HostServiceExecResponse{}, nil
}

func (s *HostGRPCServer) FindApplication(ctx context.Context, req *hyprpanelv1.HostServiceFindApplicationRequest) (*hyprpanelv1.HostServiceFindApplicationResponse, error) {
	appInfo, err := s.Impl.FindApplication(req.Query)
	if err != nil {
		return &hyprpanelv1.HostServiceFindApplicationResponse{}, err
	}

	return &hyprpanelv1.HostServiceFindApplicationResponse{
		AppInfo: appInfo,
	}, nil
}

func (s *HostGRPCServer) SystrayActivate(ctx context.Context, req *hyprpanelv1.HostServiceSystrayActivateRequest) (*hyprpanelv1.HostServiceSystrayActivateResponse, error) {
	err := s.Impl.SystrayActivate(req.BusName, req.X, req.Y)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayActivateResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayActivateResponse{}, nil
}

func (s *HostGRPCServer) SystraySecondaryActivate(ctx context.Context, req *hyprpanelv1.HostServiceSystraySecondaryActivateRequest) (*hyprpanelv1.HostServiceSystraySecondaryActivateResponse, error) {
	err := s.Impl.SystraySecondaryActivate(req.BusName, req.X, req.Y)
	if err != nil {
		return &hyprpanelv1.HostServiceSystraySecondaryActivateResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystraySecondaryActivateResponse{}, nil
}

func (s *HostGRPCServer) SystrayScroll(ctx context.Context, req *hyprpanelv1.HostServiceSystrayScrollRequest) (*hyprpanelv1.HostServiceSystrayScrollResponse, error) {
	err := s.Impl.SystrayScroll(req.BusName, req.Delta, req.Orientation)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayScrollResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayScrollResponse{}, nil
}

func (s *HostGRPCServer) SystrayMenuContextActivate(ctx context.Context, req *hyprpanelv1.HostServiceSystrayMenuContextActivateRequest) (*hyprpanelv1.HostServiceSystrayMenuContextActivateResponse, error) {
	err := s.Impl.SystrayMenuContextActivate(req.BusName, req.X, req.Y)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayMenuContextActivateResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayMenuContextActivateResponse{}, nil
}

func (s *HostGRPCServer) SystrayMenuAboutToShow(ctx context.Context, req *hyprpanelv1.HostServiceSystrayMenuAboutToShowRequest) (*hyprpanelv1.HostServiceSystrayMenuAboutToShowResponse, error) {
	err := s.Impl.SystrayMenuAboutToShow(req.BusName, req.MenuItemId)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayMenuAboutToShowResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayMenuAboutToShowResponse{}, nil
}

func (s *HostGRPCServer) SystrayMenuEvent(ctx context.Context, req *hyprpanelv1.HostServiceSystrayMenuEventRequest) (*hyprpanelv1.HostServiceSystrayMenuEventResponse, error) {
	timestamp := time.Unix(int64(req.Timestamp), 0)
	err := s.Impl.SystrayMenuEvent(req.BusName, req.Id, req.EventId, req.Data, timestamp)
	if err != nil {
		return &hyprpanelv1.HostServiceSystrayMenuEventResponse{}, err
	}

	return &hyprpanelv1.HostServiceSystrayMenuEventResponse{}, nil
}

func (s *HostGRPCServer) NotificationClosed(ctx context.Context, req *hyprpanelv1.HostServiceNotificationClosedRequest) (*hyprpanelv1.HostServiceNotificationClosedResponse, error) {
	err := s.Impl.NotificationClosed(req.Id, req.Reason)
	if err != nil {
		return &hyprpanelv1.HostServiceNotificationClosedResponse{}, err
	}

	return &hyprpanelv1.HostServiceNotificationClosedResponse{}, nil
}

func (s *HostGRPCServer) NotificationAction(ctx context.Context, req *hyprpanelv1.HostServiceNotificationActionRequest) (*hyprpanelv1.HostServiceNotificationActionResponse, error) {
	err := s.Impl.NotificationAction(req.Id, req.ActionKey)
	if err != nil {
		return &hyprpanelv1.HostServiceNotificationActionResponse{}, err
	}

	return &hyprpanelv1.HostServiceNotificationActionResponse{}, nil
}

func (s *HostGRPCServer) AudioSinkVolumeAdjust(ctx context.Context, req *hyprpanelv1.HostServiceAudioSinkVolumeAdjustRequest) (*hyprpanelv1.HostServiceAudioSinkVolumeAdjustResponse, error) {
	err := s.Impl.AudioSinkVolumeAdjust(req.Id, req.Direction)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSinkVolumeAdjustResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSinkVolumeAdjustResponse{}, nil
}

func (s *HostGRPCServer) AudioSinkMuteToggle(ctx context.Context, req *hyprpanelv1.HostServiceAudioSinkMuteToggleRequest) (*hyprpanelv1.HostServiceAudioSinkMuteToggleResponse, error) {
	err := s.Impl.AudioSinkMuteToggle(req.Id)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSinkMuteToggleResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSinkMuteToggleResponse{}, nil
}

func (s *HostGRPCServer) AudioSourceVolumeAdjust(ctx context.Context, req *hyprpanelv1.HostServiceAudioSourceVolumeAdjustRequest) (*hyprpanelv1.HostServiceAudioSourceVolumeAdjustResponse, error) {
	err := s.Impl.AudioSourceVolumeAdjust(req.Id, req.Direction)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSourceVolumeAdjustResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSourceVolumeAdjustResponse{}, nil
}

func (s *HostGRPCServer) AudioSourceMuteToggle(ctx context.Context, req *hyprpanelv1.HostServiceAudioSourceMuteToggleRequest) (*hyprpanelv1.HostServiceAudioSourceMuteToggleResponse, error) {
	err := s.Impl.AudioSourceMuteToggle(req.Id)
	if err != nil {
		return &hyprpanelv1.HostServiceAudioSourceMuteToggleResponse{}, err
	}

	return &hyprpanelv1.HostServiceAudioSourceMuteToggleResponse{}, nil
}

func (s *HostGRPCServer) BrightnessAdjust(ctx context.Context, req *hyprpanelv1.HostServiceBrightnessAdjustRequest) (*hyprpanelv1.HostServiceBrightnessAdjustResponse, error) {
	err := s.Impl.BrightnessAdjust(req.DevName, req.Direction)
	if err != nil {
		return &hyprpanelv1.HostServiceBrightnessAdjustResponse{}, err
	}

	return &hyprpanelv1.HostServiceBrightnessAdjustResponse{}, nil
}
