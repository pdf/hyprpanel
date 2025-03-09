package main

import (
	"github.com/jwijenbergh/puregotk/v4/gtk"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type spacer struct {
	*refTracker
	*api
	cfg *modulev1.Spacer

	container *gtk.Box
}

func (s *spacer) build(container *gtk.Box) error {
	s.container = gtk.NewBox(gtk.OrientationHorizontalValue, 0)
	s.AddRef(s.container.Unref)
	s.container.SetName(style.SpacerID)
	if s.orientation == gtk.OrientationHorizontalValue {
		s.container.SetSizeRequest(int(s.cfg.Size), int(s.panelCfg.Size))
		s.container.SetHexpand(s.cfg.Expand)
	} else {
		s.container.SetSizeRequest(int(s.panelCfg.Size), int(s.cfg.Size))
		s.container.SetVexpand(s.cfg.Expand)
	}
	container.Append(&s.container.Widget)
	return nil
}

func (s *spacer) close(container *gtk.Box) {
	log.Debug(`Closing module on request`, `module`, style.SpacerID)
	container.Remove(&s.container.Widget)
	s.Unref()
}

func newSpacer(cfg *modulev1.Spacer, a *api) *spacer {
	return &spacer{
		refTracker: newRefTracker(),
		api:        a,
		cfg:        cfg,
	}
}
