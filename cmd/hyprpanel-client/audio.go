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

const (
	audioSinkIconScale   = 0.9
	audioSourceIconScale = 0.4
)

type audio struct {
	*refTracker
	panel                *panel
	cfg                  *modulev1.Audio
	container            *gtk.CenterBox
	inner                *gtk.Fixed
	sinkContainer        *gtk.CenterBox
	sinkIcon             *gtk.Image
	sourceContainer      *gtk.CenterBox
	sourceIcon           *gtk.Image
	tooltip              string
	defaultSinkID        string
	defaultSinkName      string
	defaultSinkVolume    int32
	defaultSinkPercent   float64
	defaultSinkMute      bool
	defaultSourceID      string
	defaultSourceName    string
	defaultSourceVolume  int32
	defaultSourcePercent float64
	defaultSourceMute    bool
	eventCh              chan *eventv1.Event
	quitCh               chan struct{}
}

func (a *audio) update() error {
	if a.sinkIcon != nil {
		sinkIcon := a.sinkIcon
		defer sinkIcon.Unref()
		a.sinkIcon = nil
		if a.cfg.EnableSource {
			sourceIcon := a.sourceIcon
			defer sourceIcon.Unref()
			a.sourceIcon = nil
		}
	}

	var err error
	sinkIconScale := 1.0
	if a.cfg.EnableSource {
		sinkIconScale = audioSinkIconScale
	}
	sinkSize := int(float64(a.cfg.IconSize) * sinkIconScale)
	if a.defaultSinkMute {
		a.sinkIcon, err = createIcon(`audio-volume-muted`, sinkSize, a.cfg.IconSymbolic, nil)
	} else {
		switch {
		case a.defaultSinkPercent >= 1:
			a.sinkIcon, err = createIcon(`audio-volume-high`, sinkSize, a.cfg.IconSymbolic, nil)
		case a.defaultSinkPercent >= 0.5:
			a.sinkIcon, err = createIcon(`audio-volume-medium`, sinkSize, a.cfg.IconSymbolic, nil)
		case a.defaultSinkPercent > 0:
			a.sinkIcon, err = createIcon(`audio-volume-low`, sinkSize, a.cfg.IconSymbolic, nil)
		default:
			a.sinkIcon, err = createIcon(`audio-volume-muted`, sinkSize, a.cfg.IconSymbolic, nil)
		}
	}
	if err != nil {
		return err
	}
	a.sinkContainer.SetCenterWidget(&a.sinkIcon.Widget)

	if a.cfg.EnableSource {
		sourceSize := int(float64(a.cfg.IconSize) * audioSourceIconScale)
		if a.defaultSourceMute {
			a.sourceIcon, err = createIcon(`audio-input-microphone-muted`, sourceSize, a.cfg.IconSymbolic, nil)
			a.sourceContainer.AddCssClass(style.DisabledClass)
		} else {
			a.sourceContainer.RemoveCssClass(style.DisabledClass)
			switch {
			case a.defaultSourcePercent >= 1:
				a.sourceIcon, err = createIcon(`audio-input-microphone-high`, sourceSize, a.cfg.IconSymbolic, nil)
			case a.defaultSourcePercent >= 0.5:
				a.sourceIcon, err = createIcon(`audio-input-microphone-medium`, sourceSize, a.cfg.IconSymbolic, nil)
			case a.defaultSourcePercent > 0:
				a.sourceIcon, err = createIcon(`audio-input-microphone-low`, sourceSize, a.cfg.IconSymbolic, nil)
			default:
				a.sourceIcon, err = createIcon(`audio-input-microphone-muted`, sourceSize, a.cfg.IconSymbolic, nil)
			}
		}
		if err != nil {
			return err
		}
		a.sourceContainer.SetCenterWidget(&a.sourceIcon.Widget)
	}

	var tooltip strings.Builder
	tooltip.WriteString(`<span weight="bold">`)
	if a.defaultSinkMute {
		tooltip.WriteString(`Mute`)
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

	if a.cfg.EnableSource {
		tooltip.WriteString("\n")
		tooltip.WriteString(`<span weight="bold">`)
		if a.defaultSourceMute {
			tooltip.WriteString(`Mute`)
		} else {
			tooltip.WriteString(strconv.Itoa(int(a.defaultSourcePercent * 100)))
			tooltip.WriteString(`%`)
		}
		tooltip.WriteString(`</span>`)
		if a.defaultSourceName != `` {
			tooltip.WriteString(`<span> `)
			tooltip.WriteString(a.defaultSourceName)
			tooltip.WriteString(`</span>`)
		}
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
	a.inner = gtk.NewFixed()
	a.inner.SetSizeRequest(int(a.cfg.IconSize), int(a.cfg.IconSize))

	sinkIconScale := 1.0
	if a.cfg.EnableSource {
		sinkIconScale = audioSinkIconScale
	}
	sinkSize := int(float64(a.cfg.IconSize) * sinkIconScale)
	a.sinkContainer = gtk.NewCenterBox()
	a.sinkContainer.SetSizeRequest(sinkSize, sinkSize)
	a.sinkIcon, err = createIcon(`audio-volume-high`, sinkSize, a.cfg.IconSymbolic, nil)
	if err != nil {
		return err
	}
	a.sinkContainer.SetCenterWidget(&a.sinkIcon.Widget)
	a.inner.Put(&a.sinkContainer.Widget, 0, 0)

	if a.cfg.EnableSource {
		sourceSize := int(float64(a.cfg.IconSize) * audioSourceIconScale)
		sourcePos := float64(int(a.cfg.IconSize) - sourceSize)
		a.sourceContainer = gtk.NewCenterBox()
		a.sourceContainer.SetSizeRequest(sourceSize, sourceSize)
		a.sourceContainer.AddCssClass(style.OverlayClass)
		a.sourceIcon, err = createIcon(`audio-input-microphone-high`, sourceSize, a.cfg.IconSymbolic, nil)
		if err != nil {
			return err
		}
		a.sourceContainer.SetCenterWidget(&a.sourceIcon.Widget)
		a.inner.Put(&a.sourceContainer.Widget, sourcePos, sourcePos)
	}

	scrollCb := func(_ gtk.EventControllerScroll, dx, dy float64) bool {
		if dy < 0 {
			if err := a.panel.host.AudioSinkVolumeAdjust(a.defaultSinkID, eventv1.Direction_DIRECTION_UP); err != nil {
				log.Warn(`Volume adjustment failed`, `module`, style.AudioID, `err`, err)
			}
		} else {
			if err := a.panel.host.AudioSinkVolumeAdjust(a.defaultSinkID, eventv1.Direction_DIRECTION_DOWN); err != nil {
				log.Warn(`Volume adjustment failed`, `module`, style.AudioID, `err`, err)
			}
		}

		return true
	}
	a.AddRef(func() {
		unrefCallback(&scrollCb)
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
			if err := a.panel.host.AudioSinkMuteToggle(a.defaultSinkID); err != nil {
				log.Warn(`Mute toggle failed`, `module`, style.AudioID, `err`, err)
			}
		case uint(gdk.BUTTON_MIDDLE):
			if err := a.panel.host.AudioSourceMuteToggle(a.defaultSourceID); err != nil {
				log.Warn(`Mute toggle failed`, `module`, style.AudioID, `err`, err)
			}
		}
	}
	a.AddRef(func() {
		unrefCallback(&clickCb)
	})

	clickController := gtk.NewGestureClick()
	clickController.SetButton(0)
	clickController.ConnectReleased(&clickCb)
	a.container.AddController(&clickController.EventController)

	a.container.SetCenterWidget(&a.inner.Widget)
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
						defer unrefCallback(&cb)
						if !data.Default {
							return false
						}

						if data.Id != a.defaultSinkID || data.Name != a.defaultSinkName {
							a.defaultSinkID = data.Id
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
						defer unrefCallback(&cb)
						if !data.Default {
							return false
						}

						if data.Id != a.defaultSourceID || data.Name != a.defaultSourceName {
							a.defaultSourceID = data.Id
							a.defaultSourceName = data.Name
						}
						a.defaultSourcePercent = data.Percent
						a.defaultSourceVolume = data.Volume
						a.defaultSourceMute = data.Mute
						if err := a.update(); err != nil {
							log.Warn(`Failed updating`, `module`, style.AudioID, `err`, err)
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
	a.sinkIcon.Unref()
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
