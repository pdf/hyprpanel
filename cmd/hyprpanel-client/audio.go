package main

import (
	"strconv"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type audio struct {
	*refTracker
	panel              *panel
	cfg                *modulev1.Audio
	container          *gtk.CenterBox
	icon               *gtk.Image
	tooltip            string
	defaultSinkId      string
	defaultSinkName    string
	defaultSinkVolume  int32
	defaultSinkPercent float64
	defaultSinkMute    bool
	defaultSourceId    string
	eventCh            chan *eventv1.Event
	quitCh             chan struct{}
}

func (a *audio) update() error {
	if a.icon != nil {
		icon := a.icon
		defer icon.Unref()
		a.icon = nil
	}

	var err error
	if a.defaultSinkMute {
		a.icon, err = createIcon(`audio-volume-muted`, int(a.cfg.IconSize), a.cfg.IconSymbolic, nil)
	} else {
		switch {
		case a.defaultSinkPercent >= 1:
			a.icon, err = createIcon(`audio-volume-high`, int(a.cfg.IconSize), a.cfg.IconSymbolic, nil)
		case a.defaultSinkPercent >= 0.5:
			a.icon, err = createIcon(`audio-volume-medium`, int(a.cfg.IconSize), a.cfg.IconSymbolic, nil)
		case a.defaultSinkPercent > 0:
			a.icon, err = createIcon(`audio-volume-low`, int(a.cfg.IconSize), a.cfg.IconSymbolic, nil)
		default:
			a.icon, err = createIcon(`audio-volume-muted`, int(a.cfg.IconSize), a.cfg.IconSymbolic, nil)
		}
	}
	if err != nil {
		return err
	}

	a.container.SetCenterWidget(&a.icon.Widget)

	var tooltip strings.Builder
	tooltip.WriteString(`<span weight="bold">`)
	if a.defaultSinkMute {
		tooltip.WriteString(`[Muted]`)

	} else {
		tooltip.WriteString(strconv.Itoa(int(a.defaultSinkPercent * 100)))
		tooltip.WriteString(`%`)
	}
	tooltip.WriteString(`</span>`)
	if a.defaultSinkName != `` {
		tooltip.WriteString(`<span> `)
		tooltip.WriteString(a.defaultSinkName)
		tooltip.WriteString(`</span>`)
	}

	if a.tooltip != tooltip.String() {
		a.tooltip = tooltip.String()
		a.container.SetTooltipMarkup(tooltip.String())
	}

	return nil
}

func (a *audio) build(container *gtk.Box) error {
	var err error

	a.container = gtk.NewCenterBox()
	a.container.SetName(style.AudioID)
	a.container.AddCssClass(style.ModuleClass)
	a.icon, err = createIcon(`audio-volume-high`, int(a.cfg.IconSize), a.cfg.IconSymbolic, nil)
	if err != nil {
		return err
	}

	scrollCb := func(_ gtk.EventControllerScroll, dx, dy float64) bool {
		if dy < 0 {
			if err := a.panel.host.AudioSinkVolumeAdjust(a.defaultSinkId, eventv1.Direction_DIRECTION_UP); err != nil {
				log.Warn(`Volume adjustment failed`, `module`, style.AudioID, `err`, err)
			}
		} else {
			if err := a.panel.host.AudioSinkVolumeAdjust(a.defaultSinkId, eventv1.Direction_DIRECTION_DOWN); err != nil {
				log.Warn(`Volume adjustment failed`, `module`, style.AudioID, `err`, err)
			}
		}

		return true
	}
	a.AddRef(func() {
		glib.UnrefCallback(&scrollCb)
	})

	scrollController := gtk.NewEventControllerScroll(gtk.EventControllerScrollVerticalValue | gtk.EventControllerScrollDiscreteValue)
	scrollController.ConnectScroll(&scrollCb)
	a.container.AddController(&scrollController.EventController)

	clickCb := func(ctrl gtk.GestureClick, nPress int, x, y float64) {
		switch ctrl.GetCurrentButton() {
		case uint(gdk.BUTTON_PRIMARY):
			if err := a.panel.host.Exec(a.cfg.CommandMixer); err != nil {
				log.Warn(`Failed launching application`, `module`, style.AudioID, `cmd`, a.cfg.CommandMixer, `err`, err)
			}
		case uint(gdk.BUTTON_SECONDARY):
			if err := a.panel.host.AudioSinkMuteToggle(a.defaultSinkId); err != nil {
				log.Warn(`Mute toggle failed`, `module`, style.AudioID, `err`, err)
			}
		case uint(gdk.BUTTON_MIDDLE):
			if err := a.panel.host.AudioSourceMuteToggle(a.defaultSourceId); err != nil {
				log.Warn(`Mute toggle failed`, `module`, style.AudioID, `err`, err)
			}
		}
	}
	a.AddRef(func() {
		glib.UnrefCallback(&clickCb)
	})

	clickController := gtk.NewGestureClick()
	clickController.SetButton(0)
	clickController.ConnectReleased(&clickCb)
	a.container.AddController(&clickController.EventController)

	a.container.SetCenterWidget(&a.icon.Widget)

	container.Append(&a.container.Widget)

	if err := a.update(); err != nil {
		return err
	}

	go a.watch()

	return nil
}

func (a *audio) events() chan<- *eventv1.Event {
	return a.eventCh
}

func (a *audio) watch() {
	for {
		select {
		case <-a.quitCh:
			return
		default:
			select {
			case <-a.quitCh:
				return
			case evt := <-a.eventCh:
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_AUDIO_SINK_CHANGE:
					data := &eventv1.AudioSinkChangeValue{}
					if !evt.Data.MessageIs(data) {
						log.Warn(`Invalid event`, `module`, style.AudioID, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Warn(`Invalid event`, `module`, style.AudioID, `err`, err, `evt`, evt)
						continue
					}

					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer glib.UnrefCallback(&cb)
						if data.Default && (data.Id != a.defaultSinkId || data.Name != a.defaultSinkName) {
							a.defaultSinkId = data.Id
							a.defaultSinkName = data.Name
						}
						a.defaultSinkPercent = data.Percent
						a.defaultSinkVolume = data.Volume
						a.defaultSinkMute = data.Mute
						if err := a.update(); err != nil {
							log.Warn(`Failed updating`, `module`, style.AudioID, `err`, err)
						}
						return false
					}

					glib.IdleAdd(&cb, 0)
				case eventv1.EventKind_EVENT_KIND_AUDIO_SOURCE_CHANGE:
					data := &eventv1.AudioSourceChangeValue{}
					if !evt.Data.MessageIs(data) {
						log.Warn(`Invalid event`, `module`, style.AudioID, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Warn(`Invalid event`, `module`, style.AudioID, `err`, err, `evt`, evt)
						continue
					}

					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer glib.UnrefCallback(&cb)
						if data.Default && data.Id != a.defaultSourceId {
							a.defaultSourceId = data.Id
						}
						return false
					}

					glib.IdleAdd(&cb, 0)
				}
			}
		}
	}
}

func (a *audio) close(container *gtk.Box) {
	defer a.Unref()
	log.Debug(`Closing module on request`, `module`, style.AudioID)
	container.Remove(&a.container.Widget)
	a.icon.Unref()
}

func newAudio(p *panel, cfg *modulev1.Audio) *audio {
	a := &audio{
		refTracker: newRefTracker(),
		panel:      p,
		cfg:        cfg,
		eventCh:    make(chan *eventv1.Event),
		quitCh:     make(chan struct{}),
	}

	p.AddRef(func() {
		close(a.quitCh)
		close(a.eventCh)
	})

	return a
}
