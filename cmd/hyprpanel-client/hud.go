package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/jwijenbergh/puregotk/v4/pango"
	gtk4layershell "github.com/pdf/hyprpanel/internal/gtk4-layer-shell"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type hud struct {
	*refTracker
	panel   *panel
	cfg     *modulev1.Hud
	eventCh chan *eventv1.Event
	quitCh  chan struct{}

	overlay          *gtk.Window
	overlayContainer *gtk.Box

	itemData          *eventv1.HudNotificationValue
	item              *gtk.Revealer
	itemContainer     *gtk.Box
	itemIconContainer *gtk.CenterBox
	itemIcon          *gtk.Image
	itemTitle         *gtk.Label
	itemBody          *gtk.Label
	itemPercent       *gtk.Label
	itemGauge         *gtk.ProgressBar
	itemClosed        chan struct{}

	timer *time.Timer
}

func (h *hud) build(_ *gtk.Box) error {
	if err := h.buildOverlay(); err != nil {
		return err
	}

	return h.buildItem()
}

func (h *hud) buildItem() error {
	if h.itemData == nil {
		h.itemData = &eventv1.HudNotificationValue{}
	}

	h.item = gtk.NewRevealer()
	h.AddRef(h.item.Unref)
	h.item.SetRevealChild(false)
	h.item.SetTransitionType(gtk.RevealerTransitionTypeCrossfadeValue)

	revealCb := func() {
		if !h.item.GetChildRevealed() {
			h.item.Hide()
		}
	}
	h.item.ConnectSignal(`notify::child-revealed`, &revealCb)
	h.AddRef(func() {
		unrefCallback(&revealCb)
	})

	unmapCb := func(gtk.Widget) {
		h.itemClosed <- struct{}{}
		h.hideNotification()
	}
	h.item.ConnectUnmap(&unmapCb)
	h.AddRef(func() {
		unrefCallback(&unmapCb)
	})

	switch h.cfg.Position {
	case modulev1.Position_POSITION_LEFT, modulev1.Position_POSITION_TOP_LEFT, modulev1.Position_POSITION_BOTTOM_LEFT, modulev1.Position_POSITION_TOP, modulev1.Position_POSITION_BOTTOM:
		h.item.SetHalign(gtk.AlignStartValue)
	default:
		h.item.SetHalign(gtk.AlignEndValue)
	}

	h.itemContainer = gtk.NewBox(gtk.OrientationVerticalValue, 0)
	h.AddRef(h.itemContainer.Unref)
	h.itemContainer.SetHexpand(false)
	h.itemContainer.SetHalign(gtk.AlignEndValue)
	h.itemContainer.AddCssClass(style.HudNotificationClass)

	h.itemIconContainer = gtk.NewCenterBox()
	h.AddRef(h.itemIconContainer.Unref)
	h.itemIconContainer.AddCssClass(style.HudIconClass)

	h.itemTitle = gtk.NewLabel(``)
	h.AddRef(h.itemTitle.Unref)
	h.itemTitle.SetSelectable(false)
	h.itemTitle.SetWrap(false)
	h.itemTitle.SetEllipsize(pango.EllipsizeEndValue)
	h.itemTitle.SetMaxWidthChars(30)
	h.itemTitle.SetXalign(0.5)
	h.itemTitle.SetHalign(gtk.AlignCenterValue)
	h.itemTitle.SetHexpand(true)
	h.itemTitle.AddCssClass(`title-2`)
	h.itemTitle.AddCssClass(style.HudTitleClass)

	h.itemBody = gtk.NewLabel(``)
	h.AddRef(h.itemBody.Unref)
	h.itemBody.SetSelectable(false)
	h.itemBody.SetWrap(false)
	h.itemBody.SetEllipsize(pango.EllipsizeEndValue)
	h.itemBody.SetMaxWidthChars(30)
	h.itemBody.SetXalign(0.5)
	h.itemBody.SetHalign(gtk.AlignCenterValue)
	h.itemBody.SetHexpand(true)
	h.itemBody.AddCssClass(style.HudBodyClass)

	h.itemPercent = gtk.NewLabel(``)
	h.AddRef(h.itemPercent.Unref)
	h.itemPercent.SetSelectable(false)
	h.itemPercent.SetWrap(false)
	h.itemPercent.SetXalign(0.5)
	h.itemPercent.SetHalign(gtk.AlignCenterValue)
	h.itemPercent.SetHexpand(true)
	h.itemPercent.AddCssClass(style.HudPercentClass)

	h.itemContainer.Append(&h.itemIconContainer.Widget)
	h.itemContainer.Append(&h.itemPercent.Widget)
	h.itemContainer.Append(&h.itemTitle.Widget)
	h.itemContainer.Append(&h.itemBody.Widget)

	h.item.SetChild(&h.itemContainer.Widget)

	h.overlayContainer.Append(&h.item.Widget)

	return nil
}

func (h *hud) closeItem() {
	h.item.SetRevealChild(false)
	// Hack around reveal-child signal unreliability by explicitly hiding after a delay
	time.AfterFunc(500*time.Millisecond, func() {
		select {
		case <-h.itemClosed:
		default:
			if h.item.IsVisible() {
				h.item.Hide()
			}
		}
	})
}

func (h *hud) update(data *eventv1.HudNotificationValue) error {
	if data == nil {
		return nil
	}

	if data.Icon != `` && h.itemData.Icon != data.Icon {
		var err error

		if h.itemIcon != nil {
			icon := h.itemIcon
			defer icon.Unref()
			h.itemIcon = nil
		}

		if h.itemIcon, err = createIcon(data.Icon, int(h.cfg.NotificationIconSize), data.IconSymbolic, []string{`dialog-information`, `dialog-information-symbolic`, `notifications`, `notification`, `help-info`}); err == nil {
			h.itemIconContainer.SetCenterWidget(&h.itemIcon.Widget)
		}
	}

	if h.itemData.Title != data.Title {
		h.itemTitle.SetLabel(data.Title)
	}

	if h.itemData.Body != data.Body {
		h.itemBody.SetLabel(data.Body)
	}

	if data.Percent >= 0 {
		data.Percent = math.Round(data.Percent*100) / 100

		if h.itemGauge == nil {
			h.itemGauge = gtk.NewProgressBar()
			h.itemGauge.AddCssClass(style.HudGaugeClass)
			h.itemContainer.Append(&h.itemGauge.Widget)
		}

		if data.Percent != h.itemData.Percent {
			h.itemPercent.SetLabel(strconv.Itoa(int(data.Percent*100)) + `%`)

			if data.Percent > 1 {
				if data.PercentMax > 0 {
					h.itemGauge.SetInverted(true)
					h.itemGauge.SetFraction((data.Percent - 1) / (data.PercentMax - 1))
				} else {
					h.itemGauge.SetInverted(false)
					h.itemGauge.SetFraction(1.0)
				}
			} else {
				h.itemGauge.SetInverted(false)
				h.itemGauge.SetFraction(data.Percent)
			}
		}
	} else {
		if data.Percent != h.itemData.Percent {
			h.itemPercent.SetLabel(``)
		}

		if h.itemGauge != nil {
			itemGauge := h.itemGauge
			defer itemGauge.Unref()
			h.itemContainer.Remove(&itemGauge.Widget)
			h.itemGauge = nil
		}
	}

	h.itemData = data

	if !h.timer.Stop() {
		select {
		case <-h.timer.C:
		default:
		}
	}
	h.timer.Reset(h.cfg.Timeout.AsDuration())

	h.showNotification()

	go func() {
		<-h.timer.C
		h.closeItem()
	}()

	return nil
}

func (h *hud) buildOverlay() error {
	h.overlay = gtk.NewWindow()
	h.AddRef(h.overlay.Unref)
	h.overlay.SetVisible(false)
	h.overlay.SetName(style.HudID)
	h.overlay.SetName(style.HudOverlayID)
	h.overlay.SetApplication(h.panel.app)
	h.overlay.SetResizable(false)
	h.overlay.SetDecorated(false)
	h.overlay.SetDeletable(false)

	gtk4layershell.InitForWindow(h.overlay)
	gtk4layershell.SetNamespace(h.overlay, appName+`.`+style.HudOverlayID)
	gtk4layershell.SetLayer(h.overlay, gtk4layershell.LayerShellLayerOverlay)
	if h.cfg.Position == modulev1.Position_POSITION_UNSPECIFIED {
		h.cfg.Position = modulev1.Position_POSITION_TOP_RIGHT
	}
	switch h.cfg.Position {
	case modulev1.Position_POSITION_TOP_LEFT:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeTop, int(h.cfg.Margin))
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeLeft, int(h.cfg.Margin))
	case modulev1.Position_POSITION_TOP:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeTop, int(h.cfg.Margin))
	case modulev1.Position_POSITION_TOP_RIGHT:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeRight, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeTop, int(h.cfg.Margin))
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeRight, int(h.cfg.Margin))
	case modulev1.Position_POSITION_RIGHT:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeRight, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeRight, int(h.cfg.Margin))
	case modulev1.Position_POSITION_BOTTOM_RIGHT:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeBottom, true)
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeRight, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeBottom, int(h.cfg.Margin))
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeRight, int(h.cfg.Margin))
	case modulev1.Position_POSITION_BOTTOM:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeBottom, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeBottom, int(h.cfg.Margin))
	case modulev1.Position_POSITION_BOTTOM_LEFT:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeBottom, true)
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeBottom, int(h.cfg.Margin))
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeLeft, int(h.cfg.Margin))
	case modulev1.Position_POSITION_LEFT:
		gtk4layershell.SetAnchor(h.overlay, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetMargin(h.overlay, gtk4layershell.LayerShellEdgeLeft, int(h.cfg.Margin))
	default:
		return fmt.Errorf(`invalid notifications position: %s`, h.cfg.Position.String())
	}

	h.overlayContainer = gtk.NewBox(gtk.OrientationVerticalValue, int(h.cfg.Margin))
	h.AddRef(h.overlayContainer.Unref)
	h.overlayContainer.SetSizeRequest(0, 0)

	h.overlay.SetChild(&h.overlayContainer.Widget)

	go h.watch()

	return nil
}

func (h *hud) showNotification() {
	h.overlay.SetVisible(true)
	h.item.Show()
	h.item.SetRevealChild(true)
}

func (h *hud) hideNotification() {
	h.overlay.SetVisible(false)
}

func (h *hud) events() chan<- *eventv1.Event {
	return h.eventCh
}

func (h *hud) watch() {
	for {
		select {
		case <-h.quitCh:
			return
		default:
			select {
			case <-h.quitCh:
				return
			case evt := <-h.eventCh:
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_HUD_NOTIFY:
					data := &eventv1.HudNotificationValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.HudID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.HudID, `event`, evt)
						continue
					}

					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer unrefCallback(&cb)
						if err := h.update(data); err != nil {
							log.Warn(`Failed updating`, `module`, style.HudID, `err`, err)
						}
						return false
					}

					glib.IdleAdd(&cb, 0)
				}
			}
		}
	}
}

func (h *hud) close(_ *gtk.Box) {
	log.Debug(`Closing module on request`, `module`, style.HudID)
	defer h.Unref()
	h.overlay.Close()
	if h.itemIcon != nil {
		h.itemIcon.Unref()
	}
	if h.itemGauge != nil {
		h.itemGauge.Unref()
	}
}

func newHud(panel *panel, cfg *modulev1.Hud) *hud {
	h := &hud{
		refTracker: newRefTracker(),
		panel:      panel,
		cfg:        cfg,
		eventCh:    make(chan *eventv1.Event, 10),
		quitCh:     make(chan struct{}),
		itemClosed: make(chan struct{}, 1),
		timer:      time.NewTimer(1),
	}
	h.AddRef(func() {
		h.timer.Stop()
	})

	if !h.timer.Stop() {
		select {
		case <-h.timer.C:
		default:
		}
	}

	h.AddRef(func() {
		close(h.quitCh)
		close(h.eventCh)
	})

	return h
}
