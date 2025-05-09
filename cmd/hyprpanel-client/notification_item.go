package main

import (
	"time"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/jwijenbergh/puregotk/v4/pango"
	"github.com/pdf/hyprpanel/internal/dbus"
	"github.com/pdf/hyprpanel/internal/hypripc"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type notificationItem struct {
	*refTracker
	*api
	cfg      *modulev1.Notifications
	data     *eventv1.NotificationValue
	deleteFn func(uint32)
	timeout  time.Duration
	timer    *time.Timer
	closed   chan struct{}

	container *gtk.Revealer
}

func (i *notificationItem) focusWindow(addr string) error {
	return i.hypr.Dispatch(hypripc.DispatchFocusWindow, `address:`+addr)
}

func (i *notificationItem) build(container *gtk.Box) error {
	i.container = gtk.NewRevealer()
	i.AddRef(i.container.Unref)
	i.container.SetRevealChild(false)
	i.container.SetHexpand(false)

	revealCb := func() {
		select {
		case <-i.closed:
			return
		default:
		}
		if !i.container.GetChildRevealed() {
			var cb glib.SourceFunc
			cb = func(uintptr) bool {
				defer unrefCallback(&cb)
				i.container.Hide()
				return false
			}
			glib.IdleAdd(&cb, 0)
		}
	}
	i.AddRef(func() {
		unrefCallback(&revealCb)
	})
	i.container.ConnectSignal(`notify::child-revealed`, &revealCb)

	var unmapCb func(gtk.Widget)
	unmapCb = func(_ gtk.Widget) {
		select {
		case <-i.closed:
			return
		default:
		}
		defer unrefCallback(&unmapCb)
		close(i.closed)
		var cb glib.SourceFunc
		cb = func(uintptr) bool {
			defer unrefCallback(&cb)
			i.deleteFn(i.data.Id)
			return false
		}
		glib.IdleAdd(&cb, 0)
	}
	i.container.ConnectUnmap(&unmapCb)

	switch i.cfg.Position {
	case modulev1.Position_POSITION_BOTTOM_LEFT, modulev1.Position_POSITION_BOTTOM, modulev1.Position_POSITION_BOTTOM_RIGHT:
		i.container.SetTransitionType(gtk.RevealerTransitionTypeSlideUpValue)
	default:
		i.container.SetTransitionType(gtk.RevealerTransitionTypeSlideDownValue)
	}

	outer := gtk.NewBox(gtk.OrientationVerticalValue, 0)
	i.AddRef(outer.Unref)
	outer.AddCssClass(style.NotificationItemClass)
	outer.SetHexpand(false)
	outer.SetHalign(gtk.AlignEndValue)

	switch i.cfg.Position {
	case modulev1.Position_POSITION_LEFT, modulev1.Position_POSITION_TOP_LEFT, modulev1.Position_POSITION_BOTTOM_LEFT, modulev1.Position_POSITION_TOP, modulev1.Position_POSITION_BOTTOM:
		i.container.SetHalign(gtk.AlignStartValue)
	default:
		i.container.SetHalign(gtk.AlignEndValue)
	}

	inner := gtk.NewBox(gtk.OrientationHorizontalValue, 0)
	i.AddRef(inner.Unref)

	iconContainer := gtk.NewCenterBox()
	i.AddRef(iconContainer.Unref)
	iconContainer.AddCssClass(style.NotificationItemIconClass)
	iconContainer.SetVexpand(true)
	inner.Append(&iconContainer.Widget)

	hasIcon := false
	for _, hint := range i.data.Hints {
		switch dbus.NotificationHintKey(hint.Key) {
		case dbus.NotificationHintKeyImagePath, dbus.NotificationHintKeyImagePathAlt:
			if hasIcon {
				continue
			}
			v, err := eventv1.DataString(hint.Value)
			if err != nil || len(v) == 0 {
				continue
			}
			if icon, err := createIcon(v, int(i.cfg.NotificationIconSize), false, []string{`dialog-information`, `dialog-information-symbolic`, `notifications`, `notification`, `help-info`}); err == nil {
				defer icon.Unref()
				iconContainer.SetCenterWidget(&icon.Widget)
				hasIcon = true
			}
		case dbus.NotificationHintKeyImageData, dbus.NotificationHintKeyImageDataAlt, dbus.NotificationHintKeyIconDataAlt:
			if hasIcon {
				continue
			}

			v := &eventv1.NotificationValue_Pixmap{}
			if !hint.Value.MessageIs(v) {
				log.Debug(`Invalid notification icon type`, `module`, style.NotificationsID)
				continue
			}
			if err := hint.Value.UnmarshalTo(v); err != nil {
				log.Debug(`Failed decoding notification icon`, `module`, style.NotificationsID, `err`, err)
				continue
			}

			pixbuf, err := pixbufFromNotificationData(v, int(i.cfg.NotificationIconSize))
			if err != nil {
				log.Debug(`Failed encoding notification icon`, `module`, style.NotificationsID, `err`, err)
				continue
			}
			icon := gtk.NewImageFromPixbuf(pixbuf)
			icon.SetPixelSize(int(i.cfg.NotificationIconSize))
			iconContainer.SetCenterWidget(&icon.Widget)
			hasIcon = true

		}
	}

	if !hasIcon && i.data.AppIcon != `` {
		if icon, err := createIcon(i.data.AppIcon, int(i.cfg.NotificationIconSize), false, []string{`dialog-information`, `dialog-information-symbolic`, `notifications`, `notification`, `help-info`}); err == nil {
			iconContainer.SetCenterWidget(&icon.Widget)
		}
	}

	textContainer := gtk.NewBox(gtk.OrientationVerticalValue, 0)
	i.AddRef(textContainer.Unref)

	summary := gtk.NewLabel(``)
	i.AddRef(summary.Unref)
	summary.SetMarkup(i.data.Summary)
	summary.SetSelectable(true)
	summary.SetWrap(false)
	summary.SetEllipsize(pango.EllipsizeEndValue)
	summary.SetMaxWidthChars(30)
	summary.SetXalign(0.5)
	summary.AddCssClass(`title-2`)
	summary.SetHalign(gtk.AlignStartValue)
	summary.SetHexpand(true)
	summary.AddCssClass(style.NotificationItemSummaryClass)

	body := gtk.NewLabel(``)
	i.AddRef(body.Unref)
	body.SetSelectable(true)
	// Enable markup output, but beware that Thunderbird does not correclty encode their messages:
	// https://bugzilla.mozilla.org/show_bug.cgi?id=1432209
	body.SetMarkup(i.data.Body)
	body.SetVexpand(true)
	body.SetWrap(true)
	body.SetWrapMode(pango.WrapWordCharValue)
	body.SetMaxWidthChars(60)
	body.SetXalign(0.5)
	body.SetHalign(gtk.AlignStartValue)
	body.SetHexpand(true)
	body.AddCssClass(style.NotificationItemBodyClass)

	textContainer.Append(&summary.Widget)
	textContainer.Append(&body.Widget)
	inner.Append(&textContainer.Widget)
	outer.Append(&inner.Widget)

	hasDefaultAction := false
	if len(i.data.Actions) > 0 {
		actions := make([]*gtk.Widget, 0, len(i.data.Actions))

		for _, action := range i.data.Actions {
			action := action

			if action.Key == `default` {
				hasDefaultAction = true
				summary.SetSelectable(false)
				body.SetSelectable(false)
				inner.SetCursorFromName(`pointer`)

				continue
			}

			btn := gtk.NewButton()
			i.AddRef(btn.Unref)
			label := gtk.NewLabel(action.Value)
			i.AddRef(label.Unref)
			if action.Value == `` {
				label.SetLabel(action.Key)
			}
			label.SetMaxWidthChars(20)
			label.SetWrap(false)
			label.SetEllipsize(pango.EllipsizeMiddleValue)
			btn.SetHexpand(true)
			btn.SetChild(&label.Widget)
			cb := func(gtk.Button) {
				if err := i.host.NotificationAction(i.data.Id, action.Key); err != nil {
					log.Debug(`Failed submitting activation`, `module`, style.NotificationsID, `actionKey`, action.Key, `err`, err)
				}
			}
			i.AddRef(func() {
				unrefCallback(&cb)
			})
			btn.ConnectClicked(&cb)

			actions = append(actions, &btn.Widget)
		}

		if len(actions) > 0 {
			sepW := gtk.NewSeparator(gtk.OrientationHorizontalValue)
			i.AddRef(sepW.Unref)
			outer.Append(&sepW.Widget)
			actionContainer := gtk.NewBox(gtk.OrientationHorizontalValue, 0)
			i.AddRef(actionContainer.Unref)
			actionContainer.AddCssClass(style.NotificationItemActionsClass)

			for n, w := range actions {
				actionContainer.Append(w)
				if len(actions) > n+1 {
					actionSep := gtk.NewSeparator(gtk.OrientationVerticalValue)
					i.AddRef(actionSep.Unref)
					actionContainer.Append(&actionSep.Widget)
				}
			}

			outer.Append(&actionContainer.Widget)
		}
	}

	clickController := gtk.NewGestureClick()
	clickController.SetButton(0)
	clickCb := func(ctrl gtk.GestureClick, nPress int, x, y float64) {
		var closeCb glib.SourceFunc
		closeCb = func(uintptr) bool {
			defer unrefCallback(&closeCb)
			i.close()
			return false
		}
		switch ctrl.GetCurrentButton() {
		case uint(gdk.BUTTON_PRIMARY):
			if !hasDefaultAction {
				return
			}
			if err := i.host.NotificationAction(i.data.Id, `default`); err != nil {
				log.Debug(`Failed submitting activation`, `module`, style.NotificationsID, `actionKey`, `default`, `err`, err)
			}
			for _, hint := range i.data.Hints {
				if hint.Key == string(dbus.NotificationHintKeySenderPid) {
					pid, err := eventv1.DataInt64(hint.Value)
					if err != nil {
						log.Debug(`Malformed pid`, `module`, style.NotificationsID, hint.Key, hint.Value, `err`, err)
						return
					}
					clients, err := i.hypr.Clients()
					if err != nil {
						break
					}

					for _, client := range clients {
						if client.Pid == pid {
							if err := i.focusWindow(client.Address); err != nil {
								log.Debug(`Failed to focus window`, `module`, style.NotificationsID, `address`, client.Address, `err`, err)
							}
						}
					}

					break
				}
			}
			glib.IdleAdd(&closeCb, 0)
		case uint(gdk.BUTTON_MIDDLE):
			glib.IdleAdd(&closeCb, 0)
		}
	}
	i.AddRef(func() {
		unrefCallback(&clickCb)
	})
	clickController.ConnectReleased(&clickCb)
	outer.AddController(&clickController.EventController)

	motionController := gtk.NewEventControllerMotion()
	enterCallback := func(ctrl gtk.EventControllerMotion, x, y float64) {
		outer.AddCssClass(style.HoverClass)
		if !i.timer.Stop() {
			select {
			case <-i.timer.C:
			default:
			}
		}
	}
	leaveCallback := func(ctrl gtk.EventControllerMotion) {
		outer.AddCssClass(style.HoverClass)
		outer.RemoveCssClass(style.HoverClass)
		if !i.timer.Stop() {
			select {
			case <-i.timer.C:
			default:
			}
		}
		i.timer.Reset(i.timeout)
	}
	i.AddRef(func() {
		unrefCallback(&enterCallback)
		unrefCallback(&leaveCallback)
	})
	motionController.ConnectEnter(&enterCallback)
	motionController.ConnectLeave(&leaveCallback)

	outer.AddController(&motionController.EventController)

	i.container.SetChild(&outer.Widget)
	container.Append(&i.container.Widget)

	i.timer = time.NewTimer(i.timeout)
	i.AddRef(func() {
		i.timer.Stop()
	})

	go func() {
		select {
		case <-i.timer.C:
		case <-i.closed:
			if !i.timer.Stop() {
				select {
				case <-i.timer.C:
				default:
				}
			}
		}

		i.close()
	}()

	i.container.SetRevealChild(true)

	return nil
}

func (i *notificationItem) close() {
	select {
	case <-i.closed:
		return
	default:
	}
	i.container.SetRevealChild(false)
	// Hack around reveal-child signal unreliability by explicitly hiding after a delay
	time.AfterFunc(500*time.Millisecond, func() {
		select {
		case <-i.closed:
		default:
			var cb glib.SourceFunc
			cb = func(uintptr) bool {
				defer unrefCallback(&cb)
				if i.container.IsVisible() {
					i.container.Hide()
				}
				return false
			}
			glib.IdleAdd(&cb, 0)
		}
	})
}

func newNotificationItem(cfg *modulev1.Notifications, a *api, data *eventv1.NotificationValue, deleteFn func(uint32)) *notificationItem {
	i := &notificationItem{
		refTracker: newRefTracker(),
		api:        a,
		cfg:        cfg,
		data:       data,
		deleteFn:   deleteFn,
		closed:     make(chan struct{}, 1),
	}
	i.timeout = i.cfg.DefaultTimeout.AsDuration()
	if i.data.Timeout.AsDuration() > 0 {
		i.timeout = i.data.Timeout.AsDuration()
	}

	return i
}
