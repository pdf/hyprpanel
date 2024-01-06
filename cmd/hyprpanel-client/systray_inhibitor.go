package main

import (
	"sync"
	"time"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type systrayInhibitor struct {
	*refTracker
	mu      sync.RWMutex
	timeout time.Duration
	timer   *time.Timer
	inhib   bool
	enterCb func(ctrl gtk.EventControllerMotion, x, y float64)
	leaveCb func(ctrl gtk.EventControllerMotion)
}

func (s *systrayInhibitor) newController() *gtk.EventControllerMotion {
	ctrl := gtk.NewEventControllerMotion()
	ctrl.ConnectEnter(&s.enterCb)
	ctrl.ConnectLeave(&s.leaveCb)

	return ctrl
}

func (s *systrayInhibitor) inhibited() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.inhib
}

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
	s.timer.Reset(s.timeout)
	s.mu.Lock()
	s.inhib = false
	s.mu.Unlock()
}

func newSystrayInhibitor(timeout time.Duration) *systrayInhibitor {
	s := &systrayInhibitor{
		refTracker: newRefTracker(),
		timeout:    timeout,
		timer:      time.NewTimer(0),
	}
	s.uninhibit()

	s.enterCb = func(ctrl gtk.EventControllerMotion, x, y float64) {
		s.inhibit()
	}
	s.leaveCb = func(ctrl gtk.EventControllerMotion) {
		s.uninhibit()
	}

	s.AddRef(func() {
		glib.UnrefCallback(&s.enterCb)
		glib.UnrefCallback(&s.leaveCb)
	})

	return s
}
