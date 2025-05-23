package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type powerChangeCache map[string]*eventv1.PowerChangeValue

type powerChangeSort []*eventv1.PowerChangeValue

func (p powerChangeSort) Len() int {
	return len(p)
}

func (p powerChangeSort) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p powerChangeSort) Less(i, j int) bool {
	return p[i].Id < p[j].Id
}

func (p powerChangeCache) toSlice() powerChangeSort {
	res := make(powerChangeSort, len(p))
	i := 0
	for _, v := range p {
		res[i] = v
		i++
	}

	return res
}

type power struct {
	*refTracker
	*api
	cfg     *modulev1.Power
	cache   powerChangeCache
	tooltip string
	eventCh chan *eventv1.Event
	quitCh  chan struct{}

	container *gtk.CenterBox
	icon      *gtk.Image
}

func writeTooltip(evt *eventv1.PowerChangeValue, tooltip *strings.Builder) {
	_, err := fmt.Fprintf(tooltip, "<span weight=\"bold\">%d%%</span> ", evt.Percentage)
	if err != nil {
		return
	}
	if evt.Vendor != `` {
		tooltip.WriteString(evt.Vendor)
		tooltip.WriteString(` `)
	}
	if evt.Model != `` {
		tooltip.WriteString(evt.Model)
	} else {
		tooltip.WriteString(eventv1.PowerDefaultID)
	}
	if tooltip.Len() == 0 {
		tooltip.WriteString(`Unknown`)
	}
	switch evt.State {
	case eventv1.PowerState_POWER_STATE_CHARGING:
		tooltip.WriteString(` (<span style="italic">Charging`)
		timeToFull := evt.TimeToFull.AsDuration()
		if timeToFull > 0 {
			tooltip.WriteString(" - ")
			tooltip.WriteString(timeToFull.String())
			tooltip.WriteString(` until charged`)
		}
		tooltip.WriteString(`</span>)`)
	case eventv1.PowerState_POWER_STATE_DISCHARGING:
		tooltip.WriteString(` (<span style="italic">Discharging`)
		timeToEmpty := evt.TimeToEmpty.AsDuration()
		if timeToEmpty > 0 {
			tooltip.WriteString(" - ")
			tooltip.WriteString(timeToEmpty.String())
			tooltip.WriteString(` until empty`)
		}
		tooltip.WriteString(`</span>)`)
	case eventv1.PowerState_POWER_STATE_EMPTY:
		tooltip.WriteString(` (<span style="italic">Empty!</span>)`)
	case eventv1.PowerState_POWER_STATE_FULLY_CHARGED:
		tooltip.WriteString(` (<span style="italic">Fully Charged</span>)`)
	case eventv1.PowerState_POWER_STATE_PENDING_CHARGE:
		tooltip.WriteString(` (<span style="italic">Pending Charge</span>)`)
	case eventv1.PowerState_POWER_STATE_PENDING_DISCHARGE:
		tooltip.WriteString(` (<span style="italic">Pending Discharge</span>)`)
	default:
		tooltip.WriteString(` (<span style="italic">Unknown</span>)`)
	}
}

func (p *power) update(evt *eventv1.PowerChangeValue) error {
	var err error

	if evt.Id == eventv1.PowerDefaultID {
		prev, hasPrev := p.cache[evt.Id]
		p.cache[evt.Id] = evt
		if !hasPrev || prev.Icon != evt.Icon {
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
		return nil
	}

	tooltip := &strings.Builder{}
	p.cache[evt.Id] = evt
	s := p.cache.toSlice()
	sort.Sort(s)

	for i, v := range s {
		if v.Id == eventv1.PowerDefaultID || v.State == eventv1.PowerState_POWER_STATE_UNSPECIFIED {
			continue
		}

		if i > 0 && tooltip.Len() > 0 {
			tooltip.WriteString("\n")
		}
		writeTooltip(v, tooltip)
	}

	if p.tooltip != tooltip.String() {
		p.tooltip = tooltip.String()
		p.container.SetTooltipMarkup(p.tooltip)
	}

	return nil
}

func (p *power) build(container *gtk.Box) error {
	p.container = gtk.NewCenterBox()
	p.container.SetName(style.PowerID)
	p.container.AddCssClass(style.ModuleClass)

	scrollCb := func(_ gtk.EventControllerScroll, dx, dy float64) bool {
		if dy < 0 {
			if err := p.host.BrightnessAdjust(``, eventv1.Direction_DIRECTION_UP); err != nil {
				log.Warn(`Brightness adjustment failed`, `module`, style.PowerID, `err`, err)
			}
		} else {
			if err := p.host.BrightnessAdjust(``, eventv1.Direction_DIRECTION_DOWN); err != nil {
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

func newPower(cfg *modulev1.Power, a *api) *power {
	p := &power{
		refTracker: newRefTracker(),
		api:        a,
		cfg:        cfg,
		cache:      make(powerChangeCache),
		eventCh:    make(chan *eventv1.Event),
		quitCh:     make(chan struct{}),
	}

	p.AddRef(func() {
		close(p.quitCh)
		close(p.eventCh)
	})

	return p
}
