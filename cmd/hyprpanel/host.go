package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

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
	"golang.org/x/sync/errgroup"
)

const (
	clientName    = `hyprpanel-client`
	layerShellLib = `libgtk4-layer-shell.so`
	layerShellPkg = `gtk-layer-shell-0`
)

var (
	errReload   = errors.New(`reloading`)
	errDisabled = errors.New(`feature disabled`)
)

type host struct {
	cfg         *configv1.Config
	stylesheet  []byte
	log         hclog.Logger
	pluginLog   hclog.Logger
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

func (h *host) Exec(command string) error {
	if command == `` {
		return errors.New(`empty command`)
	}
	var (
		c    string
		args []string
	)
	if h.cfg.LogSubprocessesToJournal {
		c = `systemd-cat`
		args = []string{`sh`, `-c`, command}
	} else {
		c = `sh`
		args = []string{`-c`, command}
	}
	h.log.Debug(`Executing command`, `cmd`, c, `args`, args)
	cmd := exec.Command(c, args...)
	if !h.cfg.LogSubprocessesToJournal {
		cmd.Stdin = os.Stdin
		cmd.Stdin = os.Stdout
		cmd.Stderr = os.Stderr
	}
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

func (h *host) runPanel(clientPath string, layerShellPath string, prevPreload string, id string, cfg *configv1.Panel) (panelplugin.Panel, *plugin.Client, error) {
	socketDir := os.Getenv(plugin.EnvUnixSocketDir)
	if socketDir == `` {
		if runDir := os.Getenv(`XDG_RUNTIME_DIR`); runDir != `` {
			socketDir = filepath.Join(runDir, `hyprpanel`)
			if err := os.MkdirAll(socketDir, 0750); err != nil && err != os.ErrExist {
				h.log.Warn(`Could not create socket dir`, `path`, socketDir, `err`, err)
			} else {
				os.Setenv(plugin.EnvUnixSocketDir, socketDir)
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

	os.Setenv(`LD_PRELOAD`, layerShellPath)
	rpcClient, err := client.Client()
	if err != nil {
		return nil, nil, fmt.Errorf(`failed initializing client: %w`, err)
	}
	os.Setenv(`LD_PRELOAD`, prevPreload)

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

func (h *host) watch(hyprEvtCh <-chan *eventv1.Event) {
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
			case evt := <-hyprEvtCh:
				h.log.Trace(`Received hypr event`, `kind`, evt.Kind)
				for _, panel := range h.panels {
					panel.Notify(evt)
				}
			case evt := <-h.dbusEvtCh:
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
				default:
					for _, panel := range h.panels {
						panel.Notify(evt)
					}
				}
			case evt := <-h.audioEvtCh:
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
		return errors.New(`no panels configured`)
	}

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
	unstarted := make(chan struct{}, 1)
	unstarted <- struct{}{}
	defer func() {
		// Block return until client shuts down
		select {
		case <-unstarted:
		case <-errCtx.Done():
		}
	}()

	if err := h.connectDBUS(); err != nil {
		return fmt.Errorf("DBUS connection failed: %w", err)
	}
	if h.dbus != nil {
		defer h.dbus.Close()
	}

	if err := h.connectAudio(); err != nil {
		return fmt.Errorf("audio connection failed: %w", err)
	}
	if h.audio != nil {
		defer h.audio.Close()
	}

	hypr, err := hypripc.New(h.log)
	if err != nil {
		panic(err)
	}
	hyprEvtCh, cancel := hypr.Subscribe()
	defer func() {
		cancel()
		hypr.Close()
	}()
	hypr.StartEvents()

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
	<-unstarted

	go h.watch(hyprEvtCh)

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
		return nil
	}
}

func (h *host) Close() {
	h.apps.Close()
	close(h.quitCh)
}

func (h *host) connectDBUS() error {
	if h.cfg.Dbus == nil || !h.cfg.Dbus.Enabled {
		return nil
	}

	dbusClient, dbusEventCh, err := dbus.New(h.cfg.Dbus, h.log)
	if err != nil {
		return fmt.Errorf(`could not connect to DBUS session: %w`, err)
	}
	h.dbus = dbusClient
	h.dbusEvtCh = dbusEventCh

	return nil
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
	h := &host{
		cfg:         cfg,
		stylesheet:  stylesheet,
		log:         log,
		pluginLog:   log.Named(`plugin`),
		reloadCh:    make(chan struct{}),
		stopWatchCh: make(chan struct{}),
		dbusEvtCh:   make(<-chan *eventv1.Event),
		audioEvtCh:  make(<-chan *eventv1.Event),
		quitCh:      make(chan struct{}),
	}

	var err error
	h.apps, err = applications.New(log)
	if err != nil {
		return nil, err
	}

	return h, nil
}
