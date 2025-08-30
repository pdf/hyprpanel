// Package dbus provides access to DBUS APIs.
package dbus

import (
	"context"
	"embed"
	"errors"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/hashicorp/go-hclog"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
)

//go:embed interfaces
var ifaces embed.FS

var (
	errDbusNotSupported = dbus.NewError(`org.freedesktop.DBus.Error.NotSupported`, nil)
	errUnsupported      = errors.New(`unsupported`)
)

// Systray DBUS API, may return nil if Systray is disabled.
type Systray interface {
	Activate(busName string, x, y int32) error
	SecondaryActivate(busName string, x, y int32) error
	Scroll(busName string, delta int32, orientation hyprpanelv1.SystrayScrollOrientation) error
	MenuContextActivate(busName string, x, y int32) error
	MenuAboutToShow(busName string, menuItemID string) error
	MenuEvent(busName string, id int32, eventID hyprpanelv1.SystrayMenuEvent, data any, timestamp time.Time) error
}

// Notification DBUS API, may return nil if Notifications are disabled.
type Notification interface {
	Closed(id uint32, reason hyprpanelv1.NotificationClosedReason) error
	Action(id uint32, actionKey string) error
}

// Brightness DBUS API, may return nil if Brightness is disabled.
type Brightness interface {
	Adjust(devName string, direction eventv1.Direction) error
}

// Client for DBUS.
type Client struct {
	cfg             *configv1.Config_DBUS
	log             hclog.Logger
	sessionConn     *dbus.Conn
	systemConn      *dbus.Conn
	eventCh         chan *eventv1.Event
	quitCh          chan struct{}
	snw             *statusNotifierWatcher
	globalShortcuts *globalShortcuts
	notifications   *notifications
	brightness      *brightness
	power           *power
}

// Systray API.
func (c *Client) Systray() Systray {
	return c.snw
}

// Notification API.
func (c *Client) Notification() Notification {
	return c.notifications
}

// Brightness API.
func (c *Client) Brightness() Brightness {
	return c.brightness
}

// Events channel will deliver events from DBUS.
func (c *Client) Events() <-chan *eventv1.Event {
	return c.eventCh
}

// Close the client.
func (c *Client) Close() error {
	close(c.quitCh)
	if c.snw != nil {
		if err := c.snw.close(); err != nil {
			c.log.Warn(`Failed closing SNW session`, `err`, err)
		}
	}
	if c.notifications != nil {
		if err := c.notifications.close(); err != nil {
			c.log.Warn(`Failed closing Notifications session`, `err`, err)
		}
	}
	if c.globalShortcuts != nil {
		if err := c.globalShortcuts.close(); err != nil {
			c.log.Warn(`Failed closing GlobalShortcuts session`, `err`, err)
		}
	}
	return c.sessionConn.Close()
}

func (c *Client) init() error {
	return nil
}

// New instantiates a new DBUS client.
func New(cfg *configv1.Config_DBUS, logger hclog.Logger) (*Client, <-chan *eventv1.Event, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(cfg.ConnectTimeout.AsDuration()))
	defer cancel()
	sessionConn, err := connectDbusSession(ctx, cfg.ConnectInterval.AsDuration())
	if err != nil {
		return nil, nil, err
	}
	systemConn, err := connectDbusSystem(ctx, cfg.ConnectInterval.AsDuration())
	if err != nil {
		return nil, nil, err
	}

	c := &Client{
		cfg:         cfg,
		log:         logger,
		sessionConn: sessionConn,
		systemConn:  systemConn,
		eventCh:     make(chan *eventv1.Event, 10),
		quitCh:      make(chan struct{}),
	}

	if cfg.Notifications.Enabled {
		if c.notifications, err = newNotifications(sessionConn, logger, c.eventCh); err != nil {
			return nil, nil, err
		}
	}

	if cfg.Systray.Enabled {
		if c.snw, err = newStatusNotifierWatcher(sessionConn, logger, c.eventCh); err != nil {
			return nil, nil, err
		}
	}

	if cfg.Shortcuts.Enabled {
		if c.globalShortcuts, err = newGlobalShortcuts(sessionConn, logger, c.eventCh); err != nil {
			return nil, nil, err
		}
	}

	if cfg.Brightness.Enabled {
		if c.brightness, err = newBrightness(systemConn, logger, c.eventCh, cfg.Brightness); err != nil {
			return nil, nil, err
		}
	}

	if cfg.Power.Enabled {
		if c.power, err = newPower(systemConn, logger, c.eventCh, cfg.Power); err != nil {
			return nil, nil, err
		}
	}

	if err := c.init(); err != nil {
		return nil, nil, err
	}

	return c, c.eventCh, nil
}

func connectDbusSession(ctx context.Context, connectInterval time.Duration) (*dbus.Conn, error) {
	conn, err := dbus.SessionBus()
	if err == nil {
		return conn, nil
	}

	ticker := time.NewTicker(connectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ticker.C:
				conn, err = dbus.SessionBus()
				if err == nil {
					return conn, nil
				}
			}
		}
	}
}

func connectDbusSystem(ctx context.Context, connectInterval time.Duration) (*dbus.Conn, error) {
	conn, err := dbus.SystemBus()
	if err == nil {
		return conn, nil
	}

	ticker := time.NewTicker(connectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ticker.C:
				conn, err = dbus.SystemBus()
				if err == nil {
					return conn, nil
				}
			}
		}
	}
}

func isValidObjectPathChar(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') || c == '_' || c == '/'
}
