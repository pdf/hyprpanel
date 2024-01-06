package panelplugin

import (
	"context"
	"time"

	"github.com/hashicorp/go-plugin"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"google.golang.org/grpc"
)

// Compile-time check
var _ plugin.GRPCPlugin = &PanelPlugin{}

const (
	PanelPluginName = `panel`
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   `hyprpanel`,
	MagicCookieValue: `panel`,
}

var PluginMap = map[string]plugin.Plugin{
	PanelPluginName: &PanelPlugin{},
}

type Host interface {
	Exec(command string) error
	FindApplication(query string) (*hyprpanelv1.AppInfo, error)
	SystrayActivate(busName string, x, y int32) error
	SystraySecondaryActivate(busName string, x, y int32) error
	SystrayScroll(busName string, delta int32, orientation hyprpanelv1.SystrayScrollOrientation) error
	SystrayMenuContextActivate(busName string, x, y int32) error
	SystrayMenuAboutToShow(busName string, menuItemID string) error
	SystrayMenuEvent(busName string, id int32, eventId hyprpanelv1.SystrayMenuEvent, data any, timestamp time.Time) error
	NotificationClosed(id uint32, reason hyprpanelv1.NotificationClosedReason) error
	NotificationAction(id uint32, actionKey string) error
	AudioSinkVolumeAdjust(id string, direction eventv1.Direction) error
	AudioSinkMuteToggle(id string) error
	AudioSourceVolumeAdjust(id string, direction eventv1.Direction) error
	AudioSourceMuteToggle(id string) error
	BrightnessAdjust(devName string, direction eventv1.Direction) error
}

type Panel interface {
	Init(host Host, id string, loglevel configv1.LogLevel, config *configv1.Panel, stylesheet []byte) error
	Notify(evt *eventv1.Event)
	Context() context.Context
	Close()
}

type PanelPlugin struct {
	plugin.NetRPCUnsupportedPlugin

	Impl Panel
}

func (p *PanelPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	hyprpanelv1.RegisterPanelServiceServer(s, &PanelGRPCServer{
		Impl:   p.Impl,
		broker: broker,
	})
	return nil
}

func (p *PanelPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &PanelGRPCClient{
		client: hyprpanelv1.NewPanelServiceClient(c),
		broker: broker,
		ctx:    ctx,
	}, nil
}
