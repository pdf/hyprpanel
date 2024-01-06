package main

import (
	"math"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	gtk4layershell "github.com/pdf/hyprpanel/internal/gtk4-layer-shell"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type session struct {
	*refTracker
	panel *panel
	cfg   *modulev1.Session

	container *gtk.CenterBox
	overlay   *gtk.Window
}

func (s *session) build(container *gtk.Box) error {
	s.container = gtk.NewCenterBox()
	s.container.SetName(style.SessionID)
	s.container.AddCssClass(style.ModuleClass)
	icon, err := createIcon(`system-shutdown`, int(s.cfg.IconSize), s.cfg.IconSymbolic, nil)
	if err != nil {
		return err
	}
	s.container.SetCenterWidget(&icon.Widget)

	s.overlay = gtk.NewWindow()
	s.overlay.SetName(style.SessionOverlayID)
	s.overlay.Hide()
	gtk4layershell.InitForWindow(s.overlay)
	gtk4layershell.SetNamespace(s.overlay, appName+`.`+style.SessionOverlayID)
	gtk4layershell.SetLayer(s.overlay, gtk4layershell.LayerShellLayerOverlay)
	gtk4layershell.SetAnchor(s.overlay, gtk4layershell.LayerShellEdgeTop, true)
	gtk4layershell.SetAnchor(s.overlay, gtk4layershell.LayerShellEdgeLeft, true)
	gtk4layershell.SetAnchor(s.overlay, gtk4layershell.LayerShellEdgeRight, true)
	gtk4layershell.SetAnchor(s.overlay, gtk4layershell.LayerShellEdgeBottom, true)

	overlayContainer := gtk.NewCenterBox()
	overlayInner := gtk.NewBox(gtk.OrientationHorizontalValue, int(s.cfg.OverlayIconSize))

	buttonSize := int(math.Floor(float64(s.cfg.OverlayIconSize) * 1.5))

	if s.cfg.CommandLogout != `` {
		logoutIcon, err := createIcon(`system-log-out`, int(s.cfg.OverlayIconSize), s.cfg.OverlayIconSymbolic, nil)
		if err != nil {
			return err
		}
		logoutIcon.SetValign(gtk.AlignCenterValue)
		logoutIcon.SetHalign(gtk.AlignCenterValue)
		logout := gtk.NewBox(gtk.OrientationVerticalValue, 0)
		logout.SetValign(gtk.AlignCenterValue)
		logoutButton := gtk.NewButton()
		logoutButton.SetValign(gtk.AlignCenterValue)
		logoutButton.SetHalign(gtk.AlignCenterValue)
		logoutButton.SetSizeRequest(buttonSize, buttonSize)
		logoutButton.SetChild(&logoutIcon.Widget)
		logoutLabel := gtk.NewLabel(`Logout`)
		logout.Append(&logoutButton.Widget)
		logout.Append(&logoutLabel.Widget)
		logoutCb := func(_ gtk.Button) {
			s.overlay.Hide()
			if err := s.panel.host.Exec(s.cfg.CommandLogout); err != nil {
				log.Error(`Failed executing logout`, `module`, style.SessionID, `err`, err)
			}
		}
		logoutButton.ConnectClicked(&logoutCb)
		s.AddRef(func() {
			glib.UnrefCallback(&logoutCb)
		})
		overlayInner.Append(&logout.Widget)
	}

	if s.cfg.CommandReboot != `` {
		rebootIcon, err := createIcon(`system-reboot`, int(s.cfg.OverlayIconSize), s.cfg.OverlayIconSymbolic, nil)
		if err != nil {
			return err
		}
		rebootIcon.SetValign(gtk.AlignCenterValue)
		rebootIcon.SetHalign(gtk.AlignCenterValue)
		reboot := gtk.NewBox(gtk.OrientationVerticalValue, 0)
		reboot.SetValign(gtk.AlignCenterValue)
		rebootButton := gtk.NewButton()
		rebootButton.SetValign(gtk.AlignCenterValue)
		rebootButton.SetHalign(gtk.AlignCenterValue)
		rebootButton.SetSizeRequest(buttonSize, buttonSize)
		rebootButton.SetChild(&rebootIcon.Widget)
		rebootLabel := gtk.NewLabel(`Reboot`)
		reboot.Append(&rebootButton.Widget)
		reboot.Append(&rebootLabel.Widget)
		rebootCb := func(_ gtk.Button) {
			s.overlay.Hide()
			if err := s.panel.host.Exec(s.cfg.CommandReboot); err != nil {
				log.Error(`Failed executing reboot`, `module`, style.SessionID, `err`, err)
			}
		}
		rebootButton.ConnectClicked(&rebootCb)
		s.AddRef(func() {
			glib.UnrefCallback(&rebootCb)
		})
		overlayInner.Append(&reboot.Widget)
	}

	if s.cfg.CommandSuspend != `` {
		suspendIcon, err := createIcon(`system-suspend`, int(s.cfg.OverlayIconSize), s.cfg.OverlayIconSymbolic, nil)
		if err != nil {
			return err
		}
		suspendIcon.SetValign(gtk.AlignCenterValue)
		suspendIcon.SetHalign(gtk.AlignCenterValue)
		suspend := gtk.NewBox(gtk.OrientationVerticalValue, 0)
		suspend.SetValign(gtk.AlignCenterValue)
		suspendButton := gtk.NewButton()
		suspendButton.SetValign(gtk.AlignCenterValue)
		suspendButton.SetHalign(gtk.AlignCenterValue)
		suspendButton.SetSizeRequest(buttonSize, buttonSize)
		suspendButton.SetChild(&suspendIcon.Widget)
		suspendLabel := gtk.NewLabel(`Suspend`)
		suspend.Append(&suspendButton.Widget)
		suspend.Append(&suspendLabel.Widget)
		suspendCb := func(_ gtk.Button) {
			s.overlay.Hide()
			if err := s.panel.host.Exec(s.cfg.CommandSuspend); err != nil {
				log.Error(`Failed executing suspend`, `module`, style.SessionID, `err`, err)
			}
		}
		suspendButton.ConnectClicked(&suspendCb)
		s.AddRef(func() {
			glib.UnrefCallback(&suspendCb)
		})
		overlayInner.Append(&suspend.Widget)
	}

	if s.cfg.CommandShutdown != `` {
		shutdownIcon, err := createIcon(`system-shutdown`, int(s.cfg.OverlayIconSize), s.cfg.OverlayIconSymbolic, nil)
		if err != nil {
			return err
		}
		shutdownIcon.SetValign(gtk.AlignCenterValue)
		shutdownIcon.SetHalign(gtk.AlignCenterValue)
		shutdown := gtk.NewBox(gtk.OrientationVerticalValue, 0)
		shutdown.SetValign(gtk.AlignCenterValue)
		shutdownButton := gtk.NewButton()
		shutdownButton.SetValign(gtk.AlignCenterValue)
		shutdownButton.SetHalign(gtk.AlignCenterValue)
		shutdownButton.SetSizeRequest(buttonSize, buttonSize)
		shutdownButton.SetChild(&shutdownIcon.Widget)
		shutdownLabel := gtk.NewLabel(`Shutdown`)
		shutdown.Append(&shutdownButton.Widget)
		shutdown.Append(&shutdownLabel.Widget)
		shutdownCb := func(_ gtk.Button) {
			s.overlay.Hide()
			if err := s.panel.host.Exec(s.cfg.CommandShutdown); err != nil {
				log.Error(`Failed executing suspend`, `module`, style.SessionID, `err`, err)
			}
		}
		shutdownButton.ConnectClicked(&shutdownCb)
		s.AddRef(func() {
			glib.UnrefCallback(&shutdownCb)
		})
		overlayInner.Append(&shutdown.Widget)
	}

	overlayContainer.SetCenterWidget(&overlayInner.Widget)

	s.overlay.SetChild(&overlayContainer.Widget)

	overlayClickController := gtk.NewGestureClick()
	overlayClickCb := func(ctrl gtk.GestureClick, nPress int, x, y float64) {
		s.overlay.Hide()
	}
	overlayClickController.ConnectReleased(&overlayClickCb)
	s.overlay.AddController(&overlayClickController.EventController)
	s.AddRef(func() {
		glib.UnrefCallback(&overlayClickCb)
	})

	overlayKeyController := gtk.NewEventControllerKey()
	overlayKeyCb := func(_ gtk.EventControllerKey, keyVal uint, keyCode uint, mods gdk.ModifierType) {
		if keyVal == uint(gdk.KEY_Escape) {
			s.overlay.Hide()
		}
	}
	overlayKeyController.ConnectKeyReleased(&overlayKeyCb)
	s.overlay.AddController(&overlayKeyController.EventController)
	s.AddRef(func() {
		glib.UnrefCallback(&overlayKeyCb)
	})

	buttonClickController := gtk.NewGestureClick()
	buttonClickCb := func(ctrl gtk.GestureClick, nPress int, x, y float64) {
		if ctrl.GetCurrentButton() == uint(gdk.BUTTON_PRIMARY) {
			s.overlay.Show()
		}
	}
	buttonClickController.ConnectReleased(&buttonClickCb)
	s.container.AddController(&buttonClickController.EventController)
	s.AddRef(func() {
		glib.UnrefCallback(&buttonClickCb)
	})

	container.Append(&s.container.Widget)

	return nil
}

func (s *session) close(container *gtk.Box) {
	log.Debug(`Closing module on request`, `module`, style.SessionID)
	container.Remove(&s.container.Widget)
	s.Unref()
}

func newSession(panel *panel, cfg *modulev1.Session) *session {
	return &session{
		refTracker: newRefTracker(),
		panel:      panel,
		cfg:        cfg,
	}
}
