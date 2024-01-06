package main

import (
	"strings"
	"time"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type clock struct {
	*refTracker
	panel     *panel
	cfg       *modulev1.Clock
	container *gtk.Box
	timeLabel *gtk.Label
	dateLabel *gtk.Label
	revealer  *gtk.Revealer
	popover   *gtk.Popover
	tooltip   string

	updateCallback glib.SourceFunc
}

func (c *clock) build(container *gtk.Box) error {
	c.container = gtk.NewBox(c.panel.orientation, 0)
	now := time.Now()
	c.timeLabel = gtk.NewLabel(now.Format(c.cfg.TimeFormat))
	c.dateLabel = gtk.NewLabel(c.cfg.DateFormat)

	c.container.Append(&c.timeLabel.Widget)
	c.container.Append(&c.dateLabel.Widget)

	if c.panel.orientation == gtk.OrientationHorizontalValue {
		c.container.SetSizeRequest(-1, int(c.panel.cfg.Size))
		c.timeLabel.SetMarginEnd(8)
	} else {
		c.container.SetSizeRequest(int(c.panel.cfg.Size), -1)
		c.timeLabel.SetMarginBottom(8)
	}

	c.timeLabel.SetName(style.ClockTimeID)
	c.dateLabel.SetName(style.ClockDateID)
	c.container.SetName(style.ClockID)
	c.container.AddCssClass(style.ModuleClass)

	calendar := gtk.NewCalendar()
	c.AddRef(calendar.Unref)
	calendar.SetName(style.ClockCalendarID)
	c.revealer = gtk.NewRevealer()
	c.revealer.SetChild(&calendar.Widget)
	c.popover = gtk.NewPopover()
	c.popover.SetChild(&c.revealer.Widget)

	closedCb := func(_ gtk.Popover) {
		c.revealer.SetRevealChild(false)
	}
	c.popover.ConnectClosed(&closedCb)

	switch c.panel.cfg.Edge {
	case configv1.Edge_EDGE_TOP:
		c.popover.SetPosition(gtk.PosBottomValue)
		c.revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideDownValue)
	case configv1.Edge_EDGE_RIGHT:
		c.popover.SetPosition(gtk.PosLeftValue)
		c.revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideLeftValue)
	case configv1.Edge_EDGE_BOTTOM:
		c.popover.SetPosition(gtk.PosTopValue)
		c.revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideUpValue)
	case configv1.Edge_EDGE_LEFT:
		c.popover.SetPosition(gtk.PosRightValue)
		c.revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideRightValue)
	}

	c.container.Append(&c.popover.Widget)

	clickController := gtk.NewGestureClick()
	clickCb := func(ctrl gtk.GestureClick, nPress int, x, y float64) {
		if ctrl.GetCurrentButton() == uint(gdk.BUTTON_PRIMARY) {
			calendar.SelectDay(glib.NewDateTimeNowLocal())
			c.popover.Popup()
			c.revealer.SetRevealChild(true)
		}
	}
	c.AddRef(func() {
		glib.UnrefCallback(&clickCb)
	})
	clickController.ConnectReleased(&clickCb)
	c.container.AddController(&clickController.EventController)

	container.Append(&c.container.Widget)

	glib.IdleAdd(&c.updateCallback, 0)

	go c.watch()

	return nil
}

func (c *clock) update() error {
	now := time.Now()
	c.timeLabel.SetLabel(now.Format(c.cfg.TimeFormat))
	c.dateLabel.SetLabel(now.Format(c.cfg.DateFormat))
	var tooltip strings.Builder
	tooltip.WriteString(`<span weight="bold">`)
	tooltip.WriteString(now.Format(c.cfg.TooltipTimeFormat))
	tooltip.WriteString(`</span><span> `)
	tooltip.WriteString(now.Format(c.cfg.TooltipDateFormat))
	tooltip.WriteString(`</span><span style="italic"> (Local)</span>`)
	for _, region := range c.cfg.AdditionalRegions {
		tooltip.WriteString("\r")
		loc, err := time.LoadLocation(region)
		if err != nil {
			continue
		}
		rtime := now.In(loc)
		tooltip.WriteString(`<span weight="bold">`)
		tooltip.WriteString(rtime.Format(c.cfg.TooltipTimeFormat))
		tooltip.WriteString(`</span><span> `)
		tooltip.WriteString(rtime.Format(c.cfg.TooltipDateFormat))
		tooltip.WriteString(`</span><span style="italic">`)
		tooltip.WriteString(` (`)
		tooltip.WriteString(region)
		tooltip.WriteString(`)`)
		tooltip.WriteString(`</span>`)
	}
	if c.tooltip != tooltip.String() {
		c.container.SetTooltipMarkup(tooltip.String())
		c.tooltip = tooltip.String()
	}

	return nil
}

func (c *clock) watch() {
	ticker := time.NewTicker(time.Second)
	c.AddRef(ticker.Stop)

	for range ticker.C {
		glib.IdleAdd(&c.updateCallback, 0)
	}
}

func (c *clock) close(container *gtk.Box) {
	log.Debug(`Closing module on request`, `module`, style.ClockID)
	container.Remove(&c.container.Widget)
	c.Unref()
}

func newClock(p *panel, cfg *modulev1.Clock) *clock {
	c := &clock{
		refTracker: newRefTracker(),
		panel:      p,
		cfg:        cfg,
	}

	c.updateCallback = func(uintptr) bool {
		if err := c.update(); err != nil {
			log.Warn(`failed updating clock`, `err`, err)
			return false
		}
		return false
	}

	return c
}
