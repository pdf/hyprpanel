package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	gtk4layershell "github.com/pdf/hyprpanel/internal/gtk4-layer-shell"
	"github.com/pdf/hyprpanel/internal/hypripc"
	"github.com/pdf/hyprpanel/internal/panelplugin"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

const appName = `com.c0dedbad.hyprpanel.client`

var errNotFound = errors.New(`not found`)

type api struct {
	host           panelplugin.Host
	hypr           *hypripc.HyprIPC
	orientation    gtk.Orientation
	currentMonitor *hypripc.Monitor
	panelCfg       *configv1.Panel
	app            *gtk.Application
}

func (a *api) pluginHost() panelplugin.Host {
	return a.host
}

type panel struct {
	*refTracker
	*api

	id         string
	stylesheet []byte

	currentGDKMonitor *gdk.Monitor

	win       *gtk.Window
	container *gtk.Box

	modules   []module
	eventCh   chan *eventv1.Event
	receivers map[module]chan<- *eventv1.Event
	readyCh   chan struct{}
	quitCh    chan struct{}
}

func (p *panel) Init(host panelplugin.Host, id string, loglevel configv1.LogLevel, cfg *configv1.Panel, stylesheet []byte) error {
	defer close(p.readyCh)
	log.SetLevel(hclog.Level(loglevel))
	p.api.host = host
	p.id = id
	p.panelCfg = cfg
	p.stylesheet = stylesheet

	switch cfg.Edge {
	case configv1.Edge_EDGE_TOP, configv1.Edge_EDGE_BOTTOM:
		p.orientation = gtk.OrientationHorizontalValue
	default:
		p.orientation = gtk.OrientationVerticalValue
	}

	return nil
}

func (p *panel) Notify(evt *eventv1.Event) {
	select {
	case <-p.quitCh:
		return
	default:
		p.eventCh <- evt
	}
}

func (p *panel) Context() context.Context {
	return nil
}

func (p *panel) Close() {
	log.Warn(`received close request`)
	p.app.Quit()
}

func (p *panel) initWindow() error {
	display := gdk.DisplayGetDefault()
	p.AddRef(display.Unref)

	defaultCSSProvider := gtk.NewCssProvider()
	p.AddRef(defaultCSSProvider.Unref)
	defaultCSSProvider.LoadFromData(string(style.Default), len(style.Default))
	gtk.StyleContextAddProviderForDisplay(display, defaultCSSProvider, uint(gtk.STYLE_PROVIDER_PRIORITY_APPLICATION))

	if len(p.stylesheet) > 0 {
		userCSSProvider := gtk.NewCssProvider()
		p.AddRef(userCSSProvider.Unref)
		userCSSProvider.LoadFromData(string(p.stylesheet), len(p.stylesheet))
		gtk.StyleContextAddProviderForDisplay(display, userCSSProvider, uint(gtk.STYLE_PROVIDER_PRIORITY_USER))
	}

	p.win = gtk.NewWindow()
	p.AddRef(p.win.Unref)
	if p.panelCfg.Id != `` {
		p.win.SetName(p.panelCfg.Id)
	}
	p.win.SetName(style.PanelID)
	p.win.SetApplication(p.app)
	p.win.SetCanFocus(false)
	p.win.SetDecorated(false)
	p.win.SetDeletable(false)

	if p.orientation == gtk.OrientationHorizontalValue {
		p.win.SetDefaultSize(-1, int(p.panelCfg.Size))
	} else {
		p.win.SetDefaultSize(int(p.panelCfg.Size), -1)
	}
	gtk4layershell.InitForWindow(p.win)

	hyprMonitors, err := p.hypr.Monitors()
	if err != nil {
		return err
	}
	if p.panelCfg.Monitor != `` {
		for _, mon := range hyprMonitors {
			mon := mon
			if mon.Name == p.panelCfg.Monitor {
				p.currentMonitor = &mon
				break
			}
		}
	}
	if p.currentMonitor == nil {
		p.currentMonitor = &hyprMonitors[0]
	}
	p.currentGDKMonitor, err = gdkMonitorFromHypr(p.currentMonitor)
	if err != nil {
		p.currentGDKMonitor = gdk.MonitorNewFromInternalPtr(gdk.DisplayGetDefault().GetMonitors().GetItem(0))
	}
	p.AddRef(p.currentGDKMonitor.Unref)

	gtk4layershell.SetMonitor(p.win, p.currentGDKMonitor)
	gtk4layershell.SetNamespace(p.win, appName)
	gtk4layershell.AutoExclusiveZoneEnable(p.win)

	switch p.panelCfg.Edge {
	case configv1.Edge_EDGE_TOP:
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeRight, true)
	case configv1.Edge_EDGE_RIGHT:
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeRight, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeBottom, true)
	case configv1.Edge_EDGE_BOTTOM:
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeBottom, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeRight, true)
	case configv1.Edge_EDGE_LEFT:
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeLeft, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeTop, true)
		gtk4layershell.SetAnchor(p.win, gtk4layershell.LayerShellEdgeBottom, true)
	default:
		return fmt.Errorf(`panel %s missing position configuration`, p.id)
	}
	gtk4layershell.SetLayer(p.win, gtk4layershell.LayerShellLayerTop)

	destroyCb := func(_ gtk.Widget) {
		p.app.Quit()
	}
	p.win.ConnectDestroy(&destroyCb)

	return nil
}

func (p *panel) build() error {
	var (
		panelOrientation, containerOrientation gtk.Orientation
		panelCSSClass                          string
	)
	if p.orientation == gtk.OrientationHorizontalValue {
		panelOrientation = gtk.OrientationVerticalValue
		panelCSSClass = style.HorizontalClass
		containerOrientation = gtk.OrientationHorizontalValue
	} else {
		panelOrientation = gtk.OrientationHorizontalValue
		panelCSSClass = style.VerticalClass
		containerOrientation = gtk.OrientationVerticalValue
	}
	panelMain := gtk.NewBox(panelOrientation, 0)
	p.AddRef(panelMain.Unref)
	panelMain.AddCssClass(panelCSSClass)
	p.win.SetChild(&panelMain.Widget)

	switch p.panelCfg.Edge {
	case configv1.Edge_EDGE_TOP:
		panelMain.AddCssClass(style.TopClass)
	case configv1.Edge_EDGE_RIGHT:
		panelMain.AddCssClass(style.RightClass)
	case configv1.Edge_EDGE_BOTTOM:
		panelMain.AddCssClass(style.BottomClass)
	case configv1.Edge_EDGE_LEFT:
		panelMain.AddCssClass(style.LeftClass)
	}

	p.container = gtk.NewBox(containerOrientation, 0)
	p.AddRef(p.container.Unref)
	panelMain.Append(&p.container.Widget)

	for _, modCfg := range p.panelCfg.Modules {
		modCfg := modCfg
		switch modCfg.Kind.(type) {
		case *modulev1.Module_Pager:
			cfg := modCfg.GetPager()
			mod := newPager(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Taskbar:
			cfg := modCfg.GetTaskbar()
			mod := newTaskbar(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Systray:
			cfg := modCfg.GetSystray()
			mod := newSystray(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Notifications:
			cfg := modCfg.GetNotifications()
			mod := newNotifications(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Hud:
			cfg := modCfg.GetHud()
			mod := newHud(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Audio:
			cfg := modCfg.GetAudio()
			mod := newAudio(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Power:
			cfg := modCfg.GetPower()
			mod := newPower(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Clock:
			cfg := modCfg.GetClock()
			mod := newClock(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Session:
			cfg := modCfg.GetSession()
			mod := newSession(cfg, p.api)
			p.modules = append(p.modules, mod)
		case *modulev1.Module_Spacer:
			cfg := modCfg.GetSpacer()
			mod := newSpacer(cfg, p.api)
			p.modules = append(p.modules, mod)
		default:
			log.Warn(`Unhandled module config`, `module`, modCfg)
		}
	}

	for _, mod := range p.modules {
		mod := mod
		if rec, ok := mod.(moduleReceiver); ok {
			p.receivers[mod] = rec.events()
		}
		if err := mod.build(p.container); err != nil {
			return err
		}
		p.AddRef(func() {
			delete(p.receivers, mod)
			mod.close(p.container)
		})
	}

	p.AddRef(func() {
		close(p.eventCh)
	})
	go p.watch()

	return nil
}

func (p *panel) watch() {
	for evt := range p.eventCh {
		log.Trace(`received panel event`, `panelID`, p.id, `evt`, evt.Kind.String())
		for _, rec := range p.receivers {
			rec <- evt
		}
	}
}

func (p *panel) run() int {
	<-p.readyCh

	defer p.Unref()
	return p.app.Run(len(os.Args), os.Args)
}

func newPanel() (*panel, error) {
	hypr, err := hypripc.New(log)
	if err != nil {
		return nil, err
	}

	p := &panel{
		refTracker: newRefTracker(),
		api: &api{
			hypr: hypr,
			app:  gtk.NewApplication(appName, gio.GApplicationFlagsNoneValue),
		},
		modules:   make([]module, 0),
		eventCh:   make(chan *eventv1.Event, 10),
		receivers: make(map[module]chan<- *eventv1.Event),
		readyCh:   make(chan struct{}),
		quitCh:    make(chan struct{}),
	}
	p.AddRef(p.app.Unref)
	p.AddRef(func() {
		close(p.quitCh)
	})
	p.app.SetFlags(gio.GApplicationNonUniqueValue)

	var activate func(gio.Application)
	activate = func(_ gio.Application) {
		defer unrefCallback(&activate)
		if err := p.initWindow(); err != nil {
			log.Error(`failed initializing window`, `err`, err)
			p.app.Quit()
		}

		if err := p.build(); err != nil {
			log.Error(`Failed initializing window`, `err`, err)
			p.app.Quit()
		}

		p.win.Show()
	}

	p.app.ConnectActivate(&activate)

	return p, nil
}
