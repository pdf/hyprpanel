package main

import (
	"github.com/jwijenbergh/puregotk/v4/gtk"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
)

type module interface {
	build(container *gtk.Box) error
	close(container *gtk.Box)
}

type moduleReceiver interface {
	events() chan<- *eventv1.Event
}
