package dbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/hashicorp/go-hclog"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	globalShortcutsName = `org.freedesktop.portal.GlobalShortcuts`

	globalShortcutsMethodCreateSession = globalShortcutsName + `.CreateSession`
	globalShortcutsMethodBindShortcuts = globalShortcutsName + `.BindShortcuts`

	globalShortcutsSignalActivated   = globalShortcutsName + `.Activated`
	globalShortcutsSignalDeactivated = globalShortcutsName + `.Deactivated`

	shortcutPrefix = `com.c0dedbad.hyprpanel`

	shortcutAudioSinkVolumeUp     = shortcutPrefix + `.audioSinkVolumeUp`
	shortcutAudioSinkVolumeDown   = shortcutPrefix + `.audioSinkVolumeDown`
	shortcutAudioSinkMuteToggle   = shortcutPrefix + `.audioSinkMuteToggle`
	shortcutAudioSourceVolumeUp   = shortcutPrefix + `.audioSourceVolumeUp`
	shortcutAudioSourceVolumeDown = shortcutPrefix + `.audioSourceVolumeDown`
	shortcutAudioSourceMuteToggle = shortcutPrefix + `.audioSourceMuteToggle`

	shortcutBrightnessUp   = shortcutPrefix + `.brightnessUp`
	shortcutBrightnessDown = shortcutPrefix + `.brightnessDown`
)

type shortcutDefinition struct {
	ID   string
	Data map[string]dbus.Variant
}

type shortcutHandler struct {
	definition shortcutDefinition
	stop       chan struct{}
	repeat     *time.Ticker
	action     func() error
}

func (h *shortcutHandler) activate() error {
	if err := h.action(); err != nil {
		return err
	}

	delay := time.After(500 * time.Millisecond)
	select {
	case <-h.stop:
		return nil
	case <-delay:
		if h.repeat == nil {
			h.repeat = time.NewTicker(50 * time.Millisecond)
		} else {
			h.repeat.Reset(50 * time.Millisecond)
		}
		defer func() {
			h.repeat.Stop()
			select {
			case <-h.repeat.C:
			default:
			}
		}()

		for {
			select {
			case <-h.stop:
				return nil
			case <-h.repeat.C:
				if err := h.action(); err != nil {
					return err
				}
			}
		}
	}
}

func (h *shortcutHandler) deactivate() {
	h.stop <- struct{}{}
}

func newshortcutHandler(definition shortcutDefinition, action func() error) *shortcutHandler {
	h := &shortcutHandler{
		definition: definition,
		stop:       make(chan struct{}, 1),
		action:     action,
	}

	return h
}

type globalShortcuts struct {
	sync.RWMutex
	conn         *dbus.Conn
	log          hclog.Logger
	portalClient *portalClient
	handlers     map[string]*shortcutHandler
	eventCh      chan *eventv1.Event
	signals      chan *dbus.Signal
	quitCh       chan struct{}

	sessionObjectPath dbus.ObjectPath
	sessionObj        dbus.BusObject
}

func (s *globalShortcuts) createSession() error {
	sessionToken := s.portalClient.token()
	sessionRequestToken := s.portalClient.token()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sessSig, err := s.portalClient.request(ctx, globalShortcutsMethodCreateSession, sessionRequestToken, map[string]dbus.Variant{
		`handle_token`:         dbus.MakeVariant(sessionRequestToken),
		`session_handle_token`: dbus.MakeVariant(sessionToken),
	})
	if err != nil {
		return err
	}

	if len(sessSig.Body) != 2 {
		return fmt.Errorf("failed parsing GlobalShortcuts session body: %+v", sessSig.Body)
	}

	sessBody, ok := sessSig.Body[1].(map[string]dbus.Variant)
	if !ok {
		return fmt.Errorf("failed asserting GlobalShortcuts session body: %+v", sessSig.Body[1])
	}

	sessPathVar, ok := sessBody[`session_handle`]
	if !ok {
		return fmt.Errorf("failed obtaining GlobalShortcuts session path: %+v", sessPathVar)
	}

	if err := sessPathVar.Store(&s.sessionObjectPath); err != nil {
		return fmt.Errorf("failed storing GlobalShortcuts session path: %+v", sessPathVar)
	}
	s.sessionObj = s.conn.Object(portalName, s.sessionObjectPath)

	return nil
}

func (s *globalShortcuts) bindShortcuts() error {
	bindRequestToken := s.portalClient.token()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	shortcuts := make([]shortcutDefinition, len(s.handlers))
	i := 0
	for _, h := range s.handlers {
		shortcuts[i] = h.definition
		i++
	}

	_, err := s.portalClient.request(ctx, globalShortcutsMethodBindShortcuts, bindRequestToken,
		s.sessionObjectPath,
		shortcuts,
		``, // Empty parent_window, since I have absolutely no idea how to get the foreign wayland window identifier.
		map[string]dbus.Variant{
			`handle_token`: dbus.MakeVariant(bindRequestToken),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *globalShortcuts) init() error {
	s.handlers[shortcutAudioSinkVolumeUp] = newshortcutHandler(shortcutDefinition{
		ID: shortcutAudioSinkVolumeUp,
		Data: map[string]dbus.Variant{
			`description`:       dbus.MakeVariant(`Increase the volume of the default audio output device`),
			`preferred_trigger`: dbus.MakeVariant(`XF86AudioRaiseVolume`),
		},
	}, func() error {
		d := &eventv1.AudioSinkVolumeAdjust{
			Id:        eventv1.AudioDefaultSink,
			Direction: eventv1.Direction_DIRECTION_UP,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SINK_VOLUME_ADJUST,
			Data: data,
		}
		return nil
	})
	s.handlers[shortcutAudioSinkVolumeDown] = newshortcutHandler(shortcutDefinition{
		ID: shortcutAudioSinkVolumeDown,
		Data: map[string]dbus.Variant{
			`description`:       dbus.MakeVariant(`Decrease the volume of the default audio output device`),
			`preferred_trigger`: dbus.MakeVariant(`XF86AudioLowerVolume`),
		},
	}, func() error {
		d := &eventv1.AudioSinkVolumeAdjust{
			Id:        eventv1.AudioDefaultSink,
			Direction: eventv1.Direction_DIRECTION_DOWN,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SINK_VOLUME_ADJUST,
			Data: data,
		}
		return nil
	})
	s.handlers[shortcutAudioSinkMuteToggle] = newshortcutHandler(shortcutDefinition{
		ID: shortcutAudioSinkMuteToggle,
		Data: map[string]dbus.Variant{
			`description`:       dbus.MakeVariant(`Toggle the mute status of the default audio output device`),
			`preferred_trigger`: dbus.MakeVariant(`XF86AudioMute`),
		},
	}, func() error {
		d := &eventv1.AudioSinkMuteToggle{
			Id: eventv1.AudioDefaultSink,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SINK_MUTE_TOGGLE,
			Data: data,
		}
		return nil
	})
	s.handlers[shortcutAudioSourceVolumeUp] = newshortcutHandler(shortcutDefinition{
		ID: shortcutAudioSourceVolumeUp,
		Data: map[string]dbus.Variant{
			`description`: dbus.MakeVariant(`Increase the volume of the default audio input device`),
		},
	}, func() error {
		d := &eventv1.AudioSourceVolumeAdjust{
			Id:        eventv1.AudioDefaultSource,
			Direction: eventv1.Direction_DIRECTION_UP,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SOURCE_VOLUME_ADJUST,
			Data: data,
		}
		return nil
	})
	s.handlers[shortcutAudioSourceVolumeDown] = newshortcutHandler(shortcutDefinition{
		ID: shortcutAudioSourceVolumeDown,
		Data: map[string]dbus.Variant{
			`description`: dbus.MakeVariant(`Decrease the volume of the default audio input device`),
		},
	}, func() error {
		d := &eventv1.AudioSourceVolumeAdjust{
			Id:        eventv1.AudioDefaultSource,
			Direction: eventv1.Direction_DIRECTION_DOWN,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SOURCE_VOLUME_ADJUST,
			Data: data,
		}
		return nil
	})
	s.handlers[shortcutAudioSourceMuteToggle] = newshortcutHandler(shortcutDefinition{
		ID: shortcutAudioSourceMuteToggle,
		Data: map[string]dbus.Variant{
			`description`:       dbus.MakeVariant(`Toggle the mute status of the default audio input device`),
			`preferred_trigger`: dbus.MakeVariant(`XF86AudioMicMute`),
		},
	}, func() error {
		d := &eventv1.AudioSourceMuteToggle{
			Id: eventv1.AudioDefaultSource,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SOURCE_MUTE_TOGGLE,
			Data: data,
		}
		return nil
	})

	s.handlers[shortcutBrightnessUp] = newshortcutHandler(shortcutDefinition{
		ID: shortcutBrightnessUp,
		Data: map[string]dbus.Variant{
			`description`:       dbus.MakeVariant(`Increase display brightness`),
			`preferred_trigger`: dbus.MakeVariant(`XF86MonBrightnessUp`),
		},
	}, func() error {
		d := &eventv1.BrightnessAdjustValue{
			Direction: eventv1.Direction_DIRECTION_UP,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_DBUS_BRIGHTNESS_ADJUST,
			Data: data,
		}
		return nil
	})
	s.handlers[shortcutBrightnessDown] = newshortcutHandler(shortcutDefinition{
		ID: shortcutBrightnessDown,
		Data: map[string]dbus.Variant{
			`description`:       dbus.MakeVariant(`Increase display brightness`),
			`preferred_trigger`: dbus.MakeVariant(`XF86MonBrightnessDown`),
		},
	}, func() error {
		d := &eventv1.BrightnessAdjustValue{
			Direction: eventv1.Direction_DIRECTION_DOWN,
		}
		data, err := anypb.New(d)
		if err != nil {
			return err
		}
		s.eventCh <- &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_DBUS_BRIGHTNESS_ADJUST,
			Data: data,
		}
		return nil
	})

	if err := s.createSession(); err != nil {
		return err
	}

	if err := s.bindShortcuts(); err != nil {
		return err
	}

	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(portalRequestName),
		dbus.WithMatchMember(portalRequestMemberResponse),
	); err != nil {
		return err
	}

	go s.watch()

	return nil
}

func (s *globalShortcuts) processActivated(sig *dbus.Signal) error {
	if len(sig.Body) != 4 {
		return fmt.Errorf("failed parsing GlobalShortcuts trigger body: %+v", sig.Body)
	}

	shortcut, ok := sig.Body[1].(string)
	if !ok {
		return fmt.Errorf("failed asserting GlobalShortcuts trigger body: %+v", sig.Body)
	}

	handler, ok := s.handlers[shortcut]
	if !ok {
		return nil
	}

	go func() {
		if err := handler.activate(); err != nil {
			s.log.Warn(`Shortcut handler failed`, `shortcut`, shortcut, `err`, err)
		}
	}()

	return nil
}

func (s *globalShortcuts) processDeactivated(sig *dbus.Signal) error {
	if len(sig.Body) != 4 {
		return fmt.Errorf("failed parsing GlobalShortcuts trigger body: %+v", sig.Body)
	}

	shortcut, ok := sig.Body[1].(string)
	if !ok {
		return fmt.Errorf("failed asserting GlobalShortcuts trigger body: %+v", sig.Body)
	}

	handler, ok := s.handlers[shortcut]
	if !ok {
		return nil
	}

	handler.deactivate()

	return nil
}

func (s *globalShortcuts) watch() {
	for {
		select {
		case <-s.quitCh:
			return
		default:
			select {
			case <-s.quitCh:
				return
			case sig, ok := <-s.signals:
				if !ok {
					return
				}
				switch sig.Name {
				case globalShortcutsSignalActivated:
					if err := s.processActivated(sig); err != nil {
						s.log.Warn(`Failed actioning global shortcut`, `sig`, sig, `err`, err)
					}
				case globalShortcutsSignalDeactivated:
					if err := s.processDeactivated(sig); err != nil {
						s.log.Warn(`Failed actioning global shortcut`, `sig`, sig, `err`, err)
					}
				}
			}
		}
	}
}

func (s *globalShortcuts) close() error {
	if s.sessionObj != nil {
		if call := s.sessionObj.Call(portalSessionMethodClose, 0); call.Err != nil {
			return call.Err
		}
	}

	return s.portalClient.conn.Close()
}

func newGlobalShortcuts(conn *dbus.Conn, logger hclog.Logger, eventCh chan *eventv1.Event) (*globalShortcuts, error) {
	p, err := newPortalClient(conn, logger)
	if err != nil {
		return nil, err
	}

	s := &globalShortcuts{
		conn:         conn,
		log:          logger,
		portalClient: p,
		handlers:     make(map[string]*shortcutHandler),
		eventCh:      eventCh,
		signals:      make(chan *dbus.Signal),
		quitCh:       make(chan struct{}),
	}

	s.conn.Signal(s.signals)

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
