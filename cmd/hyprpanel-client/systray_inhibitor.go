package main

import (
	"sync"
	"time"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
)

type systrayInhibitor struct {
	*refTracker
	*api
	cfg       *modulev1.Systray
	mu        sync.RWMutex
	timer     *time.Timer
	hidden    chan struct{}
	inhib     bool
	revealer  *gtk.Revealer
	revealBtn *gtk.Button
	enterCb   func(ctrl gtk.EventControllerMotion, x, y float64)
	leaveCb   func(ctrl gtk.EventControllerMotion)
}

func (s *systrayInhibitor) newController() *gtk.EventControllerMotion {
	ctrl := gtk.NewEventControllerMotion()
	ctrl.ConnectEnter(&s.enterCb)
	ctrl.ConnectLeave(&s.leaveCb)

	return ctrl
}

/*
func (s *systrayInhibitor) inhibited() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.inhib
}
*/

func (s *systrayInhibitor) wait() <-chan time.Time {
	return s.timer.C
}

func (s *systrayInhibitor) inhibit() {
	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	s.mu.Lock()
	s.inhib = true
	s.mu.Unlock()
}

func (s *systrayInhibitor) uninhibit() {
	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	s.timer.Reset(s.cfg.AutoHideDelay.AsDuration())
	s.mu.Lock()
	s.inhib = false
	s.mu.Unlock()
}

func (s *systrayInhibitor) updateRevealBtn() {
	if s.orientation == gtk.OrientationHorizontalValue {
		if s.revealer.GetChildRevealed() && s.revealBtn.GetLabel() != systrayRevealLabelRight {
			s.revealBtn.SetLabel(systrayRevealLabelRight)
		} else if s.revealBtn.GetLabel() != systrayRevealLabelLeft {
			s.revealBtn.SetLabel(systrayRevealLabelLeft)
		}
	} else {
		if s.revealer.GetChildRevealed() && s.revealBtn.GetLabel() != systrayRevealLabelDown {
			s.revealBtn.SetLabel(systrayRevealLabelDown)
		} else if s.revealBtn.GetLabel() != systrayRevealLabelUp {
			s.revealBtn.SetLabel(systrayRevealLabelUp)
		}
	}
}

func (s *systrayInhibitor) build(container *gtk.Box, hiddenContainer *gtk.Widget) error {
	s.revealer = gtk.NewRevealer()
	s.AddRef(s.revealer.Unref)
	s.revealer.SetRevealChild(false)

	s.revealBtn = gtk.NewButton()
	s.AddRef(s.revealBtn.Unref)
	s.updateRevealBtn()

	revealBtnCb := func(gtk.Button) {
		s.inhibit()
		s.revealer.SetRevealChild(!s.revealer.GetRevealChild())
	}
	s.AddRef(func() {
		unrefCallback(&revealBtnCb)
	})
	s.revealBtn.ConnectClicked(&revealBtnCb)

	revealCb := func() {
		s.updateRevealBtn()
		if s.cfg.AutoHideDelay.AsDuration() == 0 {
			return
		}

		if s.revealer.GetRevealChild() {
			go func() {
				select {
				case <-s.hidden:
				case <-s.wait():
					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer unrefCallback(&cb)
						s.revealer.SetRevealChild(false)
						return false
					}
					glib.IdleAdd(&cb, 0)
				}
			}()
		} else {
			s.inhibit()
			select {
			case s.hidden <- struct{}{}:
			default:
			}
		}
	}
	s.AddRef(func() {
		unrefCallback(&revealCb)
	})
	s.revealer.ConnectSignal(`notify::child-revealed`, &revealCb)

	s.revealer.SetChild(hiddenContainer)

	container.Append(&s.revealBtn.Widget)
	container.Append(&s.revealer.Widget)

	return nil
}

func newSystrayInhibitor(cfg *modulev1.Systray, a *api) *systrayInhibitor {
	s := &systrayInhibitor{
		refTracker: newRefTracker(),
		api:        a,
		cfg:        cfg,
		timer:      time.NewTimer(0),
		hidden:     make(chan struct{}),
	}
	s.uninhibit()

	s.enterCb = func(ctrl gtk.EventControllerMotion, x, y float64) {
		s.inhibit()
	}
	s.leaveCb = func(ctrl gtk.EventControllerMotion) {
		s.uninhibit()
	}

	s.AddRef(func() {
		unrefCallback(&s.enterCb)
		unrefCallback(&s.leaveCb)
	})

	return s
}
