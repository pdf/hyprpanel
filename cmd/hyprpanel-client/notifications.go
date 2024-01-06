package main

import (
	"fmt"
	"sync"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	gtk4layershell "github.com/pdf/hyprpanel/internal/gtk4-layer-shell"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/pdf/hyprpanel/style"
)

type notifications struct {
	*refTracker
	sync.RWMutex
	panel   *panel
	cfg     *modulev1.Notifications
	eventCh chan *eventv1.Event
	quitCh  chan struct{}
	items   map[uint32]*notificationItem

	container        *gtk.CenterBox
	overlay          *gtk.Window
	overlayContainer *gtk.Box
}

func (n *notifications) build(container *gtk.Box) error {
	if err := n.buildOverlay(); err != nil {
		return err
	}

	if len(n.cfg.Persistent) == 0 {
		return nil
	}

	n.container = gtk.NewCenterBox()
	n.AddRef(n.container.Unref)
	n.container.SetName(style.NotificationsID)
	n.container.AddCssClass(style.ModuleClass)
	icon, err := createIcon(`notification`, int(n.cfg.IconSize), true, []string{`notifications`})
	if err != nil {
		return err
	}
	n.AddRef(icon.Unref)
	n.container.SetCenterWidget(&icon.Widget)

	container.Append(&n.container.Widget)

	return nil
}

func (n *notifications) buildOverlay() error {
	n.overlay = gtk.NewWindow()
	n.AddRef(n.overlay.Unref)
	n.overlay.SetName(style.NotificationsOverlayID)
	n.overlay.SetApplication(n.panel.app)
	n.overlay.SetResizable(false)
	n.overlay.SetDecorated(false)
	n.overlay.SetDeletable(false)

	gtk4layershell.InitForWindow(n.overlay)
	gtk4layershell.SetNamespace(n.overlay, appName+`.`+style.NotificationsOverlayID)
	gtk4layershell.SetLayer(n.overlay, gtk4layershell.LayerShellLayerOverlay)
	if n.cfg.Position == modulev1.Position_POSITION_UNSPECIFIED {
		n.cfg.Position = modulev1.Position_POSITION_TOP_RIGHT
	}
	switch n.cfg.Position {
	case modulev1.Position_POSITION_TOP_LEFT:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeTop, int(n.cfg.Margin))
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeLeft, int(n.cfg.Margin))
	case modulev1.Position_POSITION_TOP:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeTop, int(n.cfg.Margin))
	case modulev1.Position_POSITION_TOP_RIGHT:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeRight, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeTop, int(n.cfg.Margin))
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeRight, int(n.cfg.Margin))
	case modulev1.Position_POSITION_RIGHT:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeRight, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeRight, int(n.cfg.Margin))
	case modulev1.Position_POSITION_BOTTOM_RIGHT:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeBottom, true)
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeRight, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeBottom, int(n.cfg.Margin))
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeRight, int(n.cfg.Margin))
	case modulev1.Position_POSITION_BOTTOM:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeBottom, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeBottom, int(n.cfg.Margin))
	case modulev1.Position_POSITION_BOTTOM_LEFT:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeBottom, true)
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeBottom, int(n.cfg.Margin))
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeLeft, int(n.cfg.Margin))
	case modulev1.Position_POSITION_LEFT:
		gtk4layershell.SetAnchor(n.overlay, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetMargin(n.overlay, gtk4layershell.LayerShellEdgeLeft, int(n.cfg.Margin))
	default:
		return fmt.Errorf(`invalid notifications position: %s`, n.cfg.Position.String())
	}

	n.overlayContainer = gtk.NewBox(gtk.OrientationVerticalValue, int(n.cfg.Margin))
	n.AddRef(n.overlayContainer.Unref)
	n.overlayContainer.SetSizeRequest(0, 0)

	n.overlay.SetChild(&n.overlayContainer.Widget)

	go n.watch()

	return nil
}

func (n *notifications) addNotification(item *notificationItem) {
	n.Lock()
	defer n.Unlock()
	n.overlay.SetVisible(true)
	if err := item.build(n.overlayContainer); err != nil {
		log.Warn(`Failed building notification`, `id`, item.data.Id, `err`, err)
		return
	}
	n.items[item.data.Id] = item
}

func (n *notifications) deleteNotification(id uint32) {
	n.Lock()
	defer n.Unlock()
	item, ok := n.items[id]
	if !ok {
		log.Debug(`Received delete request for unknown notification ID`, `id`, id)
	}
	delete(n.items, id)
	defer item.Unref()

	n.overlayContainer.Remove(&item.container.Widget)

	if len(n.items) == 0 {
		n.overlay.SetVisible(false)
	}

	n.panel.host.NotificationClosed(item.data.Id, hyprpanelv1.NotificationClosedReason_NOTIFICATION_CLOSED_REASON_DISMISSED)
}

func (n *notifications) events() chan<- *eventv1.Event {
	return n.eventCh
}

func (n *notifications) watch() {
	for {
		select {
		case <-n.quitCh:
			return
		default:
			select {
			case <-n.quitCh:
				return
			case evt := <-n.eventCh:
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_DBUS_NOTIFICATION:
					data := &eventv1.NotificationValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.NotificationsID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.NotificationsID, `event`, evt)
						continue
					}

					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer glib.UnrefCallback(&cb)
						item := newNotificationItem(n, data)
						n.addNotification(item)
						return false
					}

					glib.IdleAdd(&cb, 0)
				case eventv1.EventKind_EVENT_KIND_DBUS_CLOSENOTIFICATION:
					id, err := eventv1.DataUInt32(evt.Data)
					if err != nil {
						log.Error(`Invalid event`, `module`, style.NotificationsID, `event`, evt)
						continue
					}

					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer glib.UnrefCallback(&cb)
						n.RLock()
						defer n.RUnlock()
						item, ok := n.items[id]
						if !ok {
							log.Debug(`Received close request for unknown notification`, `id`, id)
							return false
						}
						item.close()
						return false
					}

					glib.IdleAdd(&cb, 0)
				}
			}
		}
	}
}

func (n *notifications) close(container *gtk.Box) {
	log.Debug(`Closing module on request`, `module`, style.NotificationsID)
	n.overlay.Close()
	if n.container != nil {
		container.Remove(&n.container.Widget)
	}
	n.Unref()
}

func newNotifications(panel *panel, cfg *modulev1.Notifications) *notifications {
	n := &notifications{
		refTracker: newRefTracker(),
		panel:      panel,
		cfg:        cfg,
		eventCh:    make(chan *eventv1.Event, 10),
		quitCh:     make(chan struct{}),
		items:      make(map[uint32]*notificationItem),
	}
	n.AddRef(func() {
		close(n.quitCh)
		close(n.eventCh)
	})

	return n
}
