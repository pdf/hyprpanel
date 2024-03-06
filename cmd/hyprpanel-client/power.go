package main

import (
	"fmt"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type power struct {
	*refTracker
	panel   *panel
	cfg     *modulev1.Power
	data    *eventv1.PowerChangeValue
	tooltip string
	eventCh chan *eventv1.Event
	quitCh  chan struct{}

	container *gtk.CenterBox
	icon      *gtk.Image
}

func (p *power) update(evt *eventv1.PowerChangeValue) error {
	var err error

	defer func() {
		p.data = evt
	}()

	if p.data == nil || p.data.Icon != evt.Icon {
		if p.icon != nil {
			icon := p.icon
			defer icon.Unref()
			p.icon = nil
		}
		p.icon, err = createIcon(evt.Icon, int(p.cfg.IconSize), p.cfg.IconSymbolic, nil)

		if err != nil {
			return err
		}

		p.container.SetCenterWidget(&p.icon.Widget)
	}

	var tooltip strings.Builder
	switch evt.State {
	case eventv1.PowerState_POWER_STATE_CHARGING:
		tooltip.WriteString(`Charging`)
		tooltip.WriteString(fmt.Sprintf(" (%d%%)", evt.Percentage))
		timeToFull := evt.TimeToFull.AsDuration()
		if timeToFull > 0 {
			tooltip.WriteString("\r")
			tooltip.WriteString(timeToFull.String())
			tooltip.WriteString(` until charged`)
		}
	case eventv1.PowerState_POWER_STATE_DISCHARGING:
		tooltip.WriteString(`Discharging`)
		tooltip.WriteString(fmt.Sprintf(" (%d%%)", evt.Percentage))
		timeToEmpty := evt.TimeToEmpty.AsDuration()
		if timeToEmpty > 0 {
			tooltip.WriteString("\r")
			tooltip.WriteString(timeToEmpty.String())
			tooltip.WriteString(` until empty`)
		}
	case eventv1.PowerState_POWER_STATE_EMPTY:
		tooltip.WriteString(`Empty!`)
		tooltip.WriteString(fmt.Sprintf(" (%d%%)", evt.Percentage))
	case eventv1.PowerState_POWER_STATE_FULLY_CHARGED:
		tooltip.WriteString(`Fully Charged`)
		tooltip.WriteString(fmt.Sprintf(" (%d%%)", evt.Percentage))
	case eventv1.PowerState_POWER_STATE_PENDING_CHARGE:
		tooltip.WriteString(`Pending Charge`)
		tooltip.WriteString(fmt.Sprintf(" (%d%%)", evt.Percentage))
	case eventv1.PowerState_POWER_STATE_PENDING_DISCHARGE:
		tooltip.WriteString(`Pending Discharge`)
		tooltip.WriteString(fmt.Sprintf(" (%d%%)", evt.Percentage))
	default:
		tooltip.WriteString(`Unknown`)
		tooltip.WriteString(fmt.Sprintf(" (%d%%)", evt.Percentage))
	}

	if p.data == nil || p.tooltip != tooltip.String() {
		p.tooltip = tooltip.String()
		p.container.SetTooltipText(p.tooltip)
	}

	p.data = evt

	return nil
}

func (p *power) build(container *gtk.Box) error {
	p.container = gtk.NewCenterBox()
	p.container.SetName(style.PowerID)
	p.container.AddCssClass(style.ModuleClass)

	scrollCb := func(_ gtk.EventControllerScroll, dx, dy float64) bool {
		if dy < 0 {
			if err := p.panel.host.BrightnessAdjust(``, eventv1.Direction_DIRECTION_UP); err != nil {
				log.Warn(`Brightness adjustment failed`, `module`, style.PowerID, `err`, err)
			}
		} else {
			if err := p.panel.host.BrightnessAdjust(``, eventv1.Direction_DIRECTION_DOWN); err != nil {
				log.Warn(`Brightness adjustment failed`, `module`, style.PowerID, `err`, err)
			}
		}

		return true
	}
	p.AddRef(func() {
		unrefCallback(&scrollCb)
	})

	scrollController := gtk.NewEventControllerScroll(gtk.EventControllerScrollVerticalValue | gtk.EventControllerScrollDiscreteValue)
	scrollController.ConnectScroll(&scrollCb)
	p.container.AddController(&scrollController.EventController)

	container.Append(&p.container.Widget)

	go p.watch()

	return nil
}

func (p *power) events() chan<- *eventv1.Event {
	return p.eventCh
}

func (p *power) watch() {
	for {
		select {
		case <-p.quitCh:
			return
		default:
			select {
			case <-p.quitCh:
				return
			case evt := <-p.eventCh:
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_DBUS_POWER_CHANGE:
					data := &eventv1.PowerChangeValue{}
					if !evt.Data.MessageIs(data) {
						log.Warn(`Invalid event`, `module`, style.PowerID, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Warn(`Invalid event`, `module`, style.PowerID, `err`, err, `evt`, evt)
						continue
					}
					if data.Id != eventv1.PowerDefaultID {
						continue
					}

					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer unrefCallback(&cb)
						if err := p.update(data); err != nil {
							log.Warn(`Failed updating`, `module`, style.PowerID, `err`, err)
						}
						return false
					}

					glib.IdleAdd(&cb, 0)
				}
			}
		}
	}
}

func (p *power) close(container *gtk.Box) {
	defer p.Unref()
	log.Debug(`Closing module on request`, `module`, style.PowerID)
	container.Remove(&p.container.Widget)
	if p.icon != nil {
		p.icon.Unref()
	}
}

func newPower(p *panel, cfg *modulev1.Power) *power {
	a := &power{
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
