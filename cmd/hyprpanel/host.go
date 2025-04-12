package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pdf/hyprpanel/internal/applications"
	"github.com/pdf/hyprpanel/internal/audio"
	"github.com/pdf/hyprpanel/internal/dbus"
	"github.com/pdf/hyprpanel/internal/hypripc"
	"github.com/pdf/hyprpanel/internal/panelplugin"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/pdf/hyprpanel/wl"
	"golang.org/x/sync/errgroup"
)

const (
	clientName    = `hyprpanel-client`
	layerShellLib = `libgtk4-layer-shell.so`
	layerShellPkg = `gtk-layer-shell-0`
)

var (
	errReload   = fmt.Errorf(`reloading`)
	errDisabled = fmt.Errorf(`feature disabled`)
)

type host struct {
	cfg         *configv1.Config
	stylesheet  []byte
	log         hclog.Logger
	pluginLog   hclog.Logger
	wl          *wl.App
	hypr        *hypripc.HyprIPC
	hyprEvtCh   <-chan *eventv1.Event
	dbus        *dbus.Client
	dbusEvtCh   <-chan *eventv1.Event
	audio       *audio.Client
	audioEvtCh  <-chan *eventv1.Event
	apps        *applications.AppCache
	panels      []panelplugin.Panel
	reloadCh    chan struct{}
	stopWatchCh chan struct{}
	quitCh      chan struct{}
}

func (h *host) Exec(action *hyprpanelv1.AppInfo_Action) error {
	if len(action.Exec) == 0 {
		return fmt.Errorf(`empty command`)
	}
	var (
		c    string
		args []string
	)
	if len(h.cfg.LaunchWrapper) == 0 {
		c = h.cfg.LaunchWrapper[0]
		args = append(h.cfg.LaunchWrapper[1:], action.Exec...)
	} else {
		c = action.Exec[0]
		args = action.Exec[1:]
	}
	h.log.Debug(`Executing command`, `cmd`, c, `args`, args)
	cmd := exec.Command(c, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func (h *host) FindApplication(query string) (*hyprpanelv1.AppInfo, error) {
	return h.apps.Find(query), nil
}

func (h *host) SystrayActivate(busName string, x, y int32) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Systray().Activate(busName, x, y)
}

func (h *host) SystraySecondaryActivate(busName string, x, y int32) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Systray().SecondaryActivate(busName, x, y)
}

func (h *host) SystrayScroll(busName string, delta int32, orientation hyprpanelv1.SystrayScrollOrientation) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Systray().Scroll(busName, delta, orientation)
}

func (h *host) SystrayMenuContextActivate(busName string, x, y int32) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Systray().MenuContextActivate(busName, x, y)
}

func (h *host) SystrayMenuAboutToShow(busName string, menuItemID string) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Systray().MenuAboutToShow(busName, menuItemID)
}

func (h *host) SystrayMenuEvent(busName string, id int32, eventID hyprpanelv1.SystrayMenuEvent, data any, timestamp time.Time) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Systray().MenuEvent(busName, id, eventID, data, timestamp)
}

func (h *host) NotificationClosed(id uint32, reason hyprpanelv1.NotificationClosedReason) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Notification().Closed(id, reason)
}

func (h *host) NotificationAction(id uint32, actionKey string) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return errDisabled
	}
	return h.dbus.Notification().Action(id, actionKey)
}

func (h *host) AudioSinkVolumeAdjust(id string, direction eventv1.Direction) error {
	if h.cfg.Audio == nil || !h.cfg.Audio.Enabled {
		return errDisabled
	}

	return h.audio.SinkVolumeAdjust(id, direction)
}

func (h *host) AudioSinkMuteToggle(id string) error {
	if h.cfg.Audio == nil || !h.cfg.Audio.Enabled {
		return errDisabled
	}

	return h.audio.SinkMuteToggle(id)
}

func (h *host) AudioSourceVolumeAdjust(id string, direction eventv1.Direction) error {
	if h.cfg.Audio == nil || !h.cfg.Audio.Enabled {
		return errDisabled
	}

	return h.audio.SourceVolumeAdjust(id, direction)
}

func (h *host) AudioSourceMuteToggle(id string) error {
	if h.cfg.Audio == nil || !h.cfg.Audio.Enabled {
		return errDisabled
	}

	return h.audio.SourceMuteToggle(id)
}

func (h *host) BrightnessAdjust(devName string, direction eventv1.Direction) error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled || h.cfg.Dbus.Brightness == nil || !h.cfg.Dbus.Brightness.Enabled {
		return errDisabled
	}

	return h.dbus.Brightness().Adjust(devName, direction)
}

func (h *host) CaptureFrame(address uint64, width, height int32) (*hyprpanelv1.ImageNRGBA, error) {
	if h.wl == nil {
		return nil, fmt.Errorf(`wl app not available`)
	}

	if address == 0 || width == 0 || height == 0 {
		return nil, fmt.Errorf("invalid parameters: address=%d, width=%d, height=%d", address, width, height)
	}

	img, err := h.wl.CaptureFrame(address)
	if err != nil {
		return nil, err
	}

	dst := imaging.Fit(img, int(width), int(height), imaging.Box)

	return &hyprpanelv1.ImageNRGBA{
		Pixels: dst.Pix,
		Stride: uint32(dst.Stride),
		Width:  uint32(dst.Bounds().Dx()),
		Height: uint32(dst.Bounds().Dy()),
	}, nil
}

func (h *host) runPanel(clientPath string, layerShellPath string, prevPreload string, id string, cfg *configv1.Panel) (panelplugin.Panel, *plugin.Client, error) {
	socketDir := os.Getenv(plugin.EnvUnixSocketDir)
	if socketDir == `` {
		if runDir := os.Getenv(`XDG_RUNTIME_DIR`); runDir != `` {
			socketDir = filepath.Join(runDir, `hyprpanel`)
			if err := os.MkdirAll(socketDir, 0o750); err != nil && err != os.ErrExist {
				h.log.Warn(`Could not create socket dir`, `path`, socketDir, `err`, err)
			} else {
				if err := os.Setenv(plugin.EnvUnixSocketDir, socketDir); err != nil {
					h.log.Warn(`Could not set socket dir`, `env`, plugin.EnvUnixSocketDir, `path`, socketDir, `err`, err)
				}
			}
		}
	}
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:     panelplugin.Handshake,
		Plugins:             panelplugin.PluginMap,
		Cmd:                 exec.Command(clientPath),
		AllowedProtocols:    []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:              h.pluginLog.Named(id),
		Managed:             true,
		GRPCBrokerMultiplex: true,
	})

	if err := os.Setenv(`LD_PRELOAD`, layerShellPath); err != nil {
		return nil, nil, fmt.Errorf(`failed to set LD_PRELOAD: %w`, err)
	}
	rpcClient, err := client.Client()
	if err != nil {
		return nil, nil, fmt.Errorf(`failed initializing client: %w`, err)
	}
	if err := os.Setenv(`LD_PRELOAD`, prevPreload); err != nil {
		return nil, nil, fmt.Errorf(`failed to restore LD_PRELOAD: %w`, err)
	}

	raw, err := rpcClient.Dispense(panelplugin.PanelPluginName)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed dispensing client: %w`, err)
	}

	panel := raw.(panelplugin.Panel)
	if err := panel.Init(h, id, h.cfg.LogLevel, cfg, h.stylesheet); err != nil {
		return nil, nil, err
	}

	return panel, client, nil
}

func (h *host) updateConfig(cfg *configv1.Config) {
	h.cfg = cfg
	h.reloadCh <- struct{}{}
}

func (h *host) updateStyle(stylesheet []byte) {
	h.stylesheet = stylesheet
	h.reloadCh <- struct{}{}
}

func (h *host) watch() {
	for {
		select {
		case <-h.stopWatchCh:
			return
		case <-h.quitCh:
			return
		default:
			select {
			case <-h.stopWatchCh:
				return
			case <-h.quitCh:
				return
			case evt, ok := <-h.hyprEvtCh:
				if !ok || evt == nil {
					h.log.Error(`Received from closed hypr event channel`)
					return
				}
				h.log.Trace(`Received hypr event`, `kind`, evt.Kind)
				for _, panel := range h.panels {
					panel.Notify(evt)
				}
			case evt, ok := <-h.dbusEvtCh:
				if !ok || evt == nil {
					h.log.Error(`Received from closed dbus event channel`)
					return
				}
				h.log.Trace(`Received dbus event`, `kind`, evt.Kind)
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_AUDIO_SINK_VOLUME_ADJUST:
					data := &eventv1.AudioSinkVolumeAdjust{}
					if !evt.Data.MessageIs(data) {
						h.log.Warn(`Invalid event`, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						h.log.Warn(`Invalid event`, `evt`, evt, `err`, err)
						continue
					}
					if err := h.AudioSinkVolumeAdjust(data.Id, data.Direction); err != nil {
						h.log.Warn(`Audio sink volume adjustment failed`, `err`, err)
					}
				case eventv1.EventKind_EVENT_KIND_AUDIO_SINK_MUTE_TOGGLE:
					data := &eventv1.AudioSinkMuteToggle{}
					if !evt.Data.MessageIs(data) {
						h.log.Warn(`Invalid event`, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						h.log.Warn(`Invalid event`, `evt`, evt, `err`, err)
						continue
					}
					if err := h.AudioSinkMuteToggle(data.Id); err != nil {
						h.log.Warn(`Audio sink mute toggle failed`, `err`, err)
					}
				case eventv1.EventKind_EVENT_KIND_AUDIO_SOURCE_VOLUME_ADJUST:
					data := &eventv1.AudioSourceVolumeAdjust{}
					if !evt.Data.MessageIs(data) {
						h.log.Warn(`Invalid event`, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						h.log.Warn(`Invalid event`, `evt`, evt, `err`, err)
						continue
					}
					if err := h.AudioSourceVolumeAdjust(data.Id, data.Direction); err != nil {
						h.log.Warn(`Audio source volume adjustment failed`, `err`, err)
					}
				case eventv1.EventKind_EVENT_KIND_AUDIO_SOURCE_MUTE_TOGGLE:
					data := &eventv1.AudioSourceMuteToggle{}
					if !evt.Data.MessageIs(data) {
						h.log.Warn(`Invalid event`, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						h.log.Warn(`Invalid event`, `evt`, evt, `err`, err)
						continue
					}
					if err := h.AudioSourceMuteToggle(data.Id); err != nil {
						h.log.Warn(`Audio source mute toggle failed`, `err`, err)
					}
				case eventv1.EventKind_EVENT_KIND_DBUS_BRIGHTNESS_ADJUST:
					data := &eventv1.BrightnessAdjustValue{}
					if !evt.Data.MessageIs(data) {
						h.log.Warn(`Invalid event`, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						h.log.Warn(`Invalid event`, `evt`, evt, `err`, err)
						continue
					}
					if err := h.BrightnessAdjust(data.DevName, data.Direction); err != nil {
						h.log.Warn(`Brightness adjustment failed`, `err`, err)
					}
				case eventv1.EventKind_EVENT_KIND_EXEC:
					data := &hyprpanelv1.AppInfo_Action{}
					if !evt.Data.MessageIs(data) {
						h.log.Warn(`Invalid event`, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						h.log.Warn(`Invalid event`, `evt`, evt, `err`, err)
						continue
					}
					if err := h.Exec(data); err != nil {
						h.log.Warn(`Exec failed`, `err`, err)
					}
				default:
					for _, panel := range h.panels {
						panel.Notify(evt)
					}
				}
			case evt, ok := <-h.audioEvtCh:
				if !ok || evt == nil {
					h.log.Error(`Received from closed audio event channel`)
					return
				}
				h.log.Trace(`Received audio event`, `kind`, evt.Kind)
				for _, panel := range h.panels {
					panel.Notify(evt)
				}
			}
		}
	}
}

func (h *host) run() error {
	if len(h.cfg.Panels) == 0 {
		return fmt.Errorf(`no panels configured`)
	}

	hyprCancel, err := h.connectHypr()
	if err != nil {
		return fmt.Errorf("hypr connection failed: %w", err)
	}
	defer func() {
		hyprCancel()
		h.hypr.Close()
	}()

	h.log.SetLevel(hclog.Level(h.cfg.LogLevel))
	h.pluginLog.SetLevel(hclog.Level(h.cfg.LogLevel))

	clientPath, err := findClient()
	if err != nil {
		return fmt.Errorf("could not find client path: %w", err)
	}

	layerShellPath, err := findLayerShell()
	if err != nil {
		return fmt.Errorf("could not find gtk4-layer-shell path: %w", err)
	}

	grp, errCtx := errgroup.WithContext(context.Background())

	if h.apps != nil {
		if err := h.apps.Close(); err != nil {
			h.log.Error(`Failed to close app cache`, `err`, err)
		}
	}

	apps, err := applications.New(h.log, h.cfg.IconOverrides)
	if err != nil {
		return fmt.Errorf("app cache initialization failed: %w", err)
	}
	h.apps = apps

	if err := h.connectDBUS(); err != nil {
		return fmt.Errorf("DBUS connection failed: %w", err)
	}
	if h.dbus != nil {
		defer func() {
			if err := h.dbus.Close(); err != nil {
				h.log.Error(`Failed to close dbus client`, `err`, err)
			}
		}()
	}

	if err := h.connectAudio(); err != nil {
		return fmt.Errorf("audio connection failed: %w", err)
	}
	if h.audio != nil {
		defer func() {
			if err := h.audio.Close(); err != nil {
				h.log.Error(`Failed to close audio client`, `err`, err)
			}
		}()
	}

	prevPreload := os.Getenv(`LD_PRELOAD`)
	h.panels = make([]panelplugin.Panel, len(h.cfg.Panels))
	for i := range h.cfg.Panels {
		cfg := h.cfg.Panels[i]

		panel, _, err := h.runPanel(clientPath, layerShellPath, prevPreload, cfg.Id, cfg)
		if err != nil {
			return fmt.Errorf("panel %s initialization failed: %w", cfg.Id, err)
		}
		h.panels[i] = panel
		defer panel.Close()

		ctx := panel.Context()
		grp.Go(func() error {
			<-ctx.Done()
			return fmt.Errorf("panel %s failed: %w", cfg.Id, panel.Context().Err())
		})
	}

	go h.watch()

	select {
	case <-h.reloadCh:
		h.stopWatchCh <- struct{}{}
		for _, panel := range h.panels {
			panel.Close()
		}
		h.panels = make([]panelplugin.Panel, 0)
		return errReload
	case <-errCtx.Done():
		select {
		case <-h.reloadCh:
			return errReload
		default:
			return fmt.Errorf("panel failed (%w): %w", grp.Wait(), errCtx.Err())
		}
	case <-h.quitCh:
		if err := h.wl.Close(); err != nil {
			h.log.Error(`Failed to close wl app`, `err`, err)
		}
		return nil
	}
}

func (h *host) Close() {
	if err := h.apps.Close(); err != nil {
		h.log.Error(`Failed to close app cache`, `err`, err)
	}
	close(h.quitCh)
}

func (h *host) connectHypr() (hypripc.CancelFunc, error) {
	var err error
	h.hypr, err = hypripc.New(h.log)
	if err != nil {
		panic(err)
	}
	var cancel hypripc.CancelFunc
	h.hyprEvtCh, cancel = h.hypr.Subscribe()
	h.hypr.StartEvents()

	return cancel, nil
}

func (h *host) connectDBUS() error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return nil
	}

	var err error
	h.dbus, h.dbusEvtCh, err = dbus.New(h.cfg.Dbus, h.log)

	return err
}

func (h *host) connectAudio() error {
	if h.cfg.Audio == nil || !h.cfg.Audio.Enabled {
		return nil
	}

	var err error
	h.audio, h.audioEvtCh, err = audio.New(h.cfg.Audio, h.log)

	return err
}

func newHost(cfg *configv1.Config, stylesheet []byte, log hclog.Logger) (*host, error) {
	var err error
	wlApp, err := wl.NewApp(log)
	if err != nil {
		return nil, err
	}

	h := &host{
		cfg:         cfg,
		stylesheet:  stylesheet,
		log:         log,
		pluginLog:   log.Named(`plugin`),
		wl:          wlApp,
		reloadCh:    make(chan struct{}),
		stopWatchCh: make(chan struct{}),
		quitCh:      make(chan struct{}),
	}

	return h, nil
}
