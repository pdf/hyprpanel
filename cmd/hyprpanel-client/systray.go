package main

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

const (
	systrayRevealLabelUp    = `⏶`
	systrayRevealLabelDown  = `⏷`
	systrayRevealLabelLeft  = `⏴`
	systrayRevealLabelRight = `⏵`
)

type systray struct {
	*refTracker
	panel     *panel
	cfg       *modulev1.Systray
	items     map[string]*systrayItem
	modules   []module
	receivers map[module]chan<- *eventv1.Event
	eventCh   chan *eventv1.Event
	hidden    chan struct{}
	quitCh    chan struct{}

	container             *gtk.Box
	clientContainer       *gtk.FlowBox
	clientHiddenContainer *gtk.FlowBox
	revealer              *gtk.Revealer
	revealBtn             *gtk.Button
	hideInhibitor         *systrayInhibitor
}

func (s *systray) addItem(itemData *eventv1.StatusNotifierValue) error {
	if _, exists := s.items[itemData.BusName]; exists {
		return errors.New(`item already exists in addItem`)
	}

	item := newSystrayItem(s, itemData)
	if err := item.build(s.clientContainer, s.clientHiddenContainer); err != nil {
		return err
	}
	s.items[itemData.BusName] = item

	if !item.pinned {
		item.autoHide(s.clientContainer, s.clientHiddenContainer)
	}

	return nil
}

func (s *systray) deleteItem(busName string) error {
	item, ok := s.items[busName]
	if !ok {
		return fmt.Errorf(`client not found: %s`, busName)
	}

	if item.hidden {
		item.close(s.clientHiddenContainer)
	} else {
		item.close(s.clientContainer)
	}
	delete(s.items, busName)
	item.Unref()

	return nil
}

func (s *systray) updateRevealBtn() {
	if s.panel.orientation == gtk.OrientationHorizontalValue {
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

func (s *systray) build(container *gtk.Box) error {
	s.container = gtk.NewBox(s.panel.orientation, 0)
	s.AddRef(s.container.Unref)
	s.container.SetName(style.SystrayID)
	s.container.AddCssClass(style.ModuleClass)

	s.revealer = gtk.NewRevealer()
	s.AddRef(s.revealer.Unref)
	s.revealer.SetRevealChild(false)

	s.clientContainer = gtk.NewFlowBox()
	s.clientContainer.SetHomogeneous(true)
	s.clientContainer.SetSelectionMode(gtk.SelectionNoneValue)
	s.clientContainer.SetActivateOnSingleClick(false)
	if s.panel.orientation == gtk.OrientationHorizontalValue {
		s.clientContainer.SetOrientation(gtk.OrientationVerticalValue)
	} else {
		s.clientContainer.SetOrientation(gtk.OrientationHorizontalValue)
	}

	s.clientHiddenContainer = gtk.NewFlowBox()
	s.AddRef(s.clientHiddenContainer.Unref)
	s.clientHiddenContainer.SetHomogeneous(true)
	s.clientHiddenContainer.SetSelectionMode(gtk.SelectionNoneValue)
	s.clientHiddenContainer.SetActivateOnSingleClick(false)
	if s.panel.orientation == gtk.OrientationHorizontalValue {
		s.clientHiddenContainer.SetOrientation(gtk.OrientationVerticalValue)
	} else {
		s.clientHiddenContainer.SetOrientation(gtk.OrientationHorizontalValue)
	}

	if s.cfg.AutoHideDelay.AsDuration() != 0 {
		hideInhibController := s.hideInhibitor.newController()
		s.container.AddController(&hideInhibController.EventController)
	}

	s.revealBtn = gtk.NewButton()
	s.AddRef(s.revealBtn.Unref)
	s.updateRevealBtn()

	revealBtnCb := func(gtk.Button) {
		s.hideInhibitor.inhibit()
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
				case <-s.hideInhibitor.wait():
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
			s.hideInhibitor.inhibit()
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

	s.revealer.SetChild(&s.clientHiddenContainer.Widget)

	if s.panel.orientation == gtk.OrientationHorizontalValue {
		s.container.SetSizeRequest(-1, int(s.panel.cfg.Size))
		s.clientContainer.SetSizeRequest(-1, int(s.panel.cfg.Size))
	} else {
		s.container.SetSizeRequest(int(s.panel.cfg.Size), -1)
		s.clientContainer.SetSizeRequest(int(s.panel.cfg.Size), -1)
	}

	for _, modCfg := range s.cfg.Modules {
		switch modCfg.Kind.(type) {
		case *modulev1.SystrayModule_Audio:
			cfg := modCfg.GetAudio()
			mod := newAudio(s.panel, cfg)
			s.modules = append(s.modules, mod)
		case *modulev1.SystrayModule_Power:
			cfg := modCfg.GetPower()
			mod := newPower(s.panel, cfg)
			s.modules = append(s.modules, mod)
		default:
			return errors.New(`unsupported systray module`)
		}
	}

	for _, mod := range s.modules {
		if rec, ok := mod.(moduleReceiver); ok {
			s.receivers[mod] = rec.events()
		}
		modContainer := gtk.NewBox(gtk.OrientationHorizontalValue, 0)
		s.AddRef(modContainer.Unref)
		modContainer.SetHalign(gtk.AlignCenterValue)
		modContainer.SetValign(gtk.AlignCenterValue)
		modContainer.SetCanFocus(false)
		modContainer.SetFocusOnClick(false)
		if err := mod.build(modContainer); err != nil {
			return err
		}
		s.AddRef(func() {
			delete(s.receivers, mod)
			mod.close(modContainer)
		})
		s.clientContainer.Append(&modContainer.Widget)
	}

	s.container.Append(&s.revealBtn.Widget)
	s.container.Append(&s.revealer.Widget)
	s.container.Append(&s.clientContainer.Widget)

	container.Append(&s.container.Widget)

	go s.watch()

	return nil
}

func (s *systray) events() chan<- *eventv1.Event {
	return s.eventCh
}

func (s *systray) watch() {
	for {
		select {
		case <-s.quitCh:
			return
		default:
			select {
			case <-s.quitCh:
				return
			case evt := <-s.eventCh:
				for _, rec := range s.receivers {
					rec <- evt
				}
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_DBUS_REGISTERSTATUSNOTIFIER:
					data := &eventv1.StatusNotifierValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					log.Trace(`Adding item`, `module`, style.SystrayID, `busName`, data.BusName)

					var addCb glib.SourceFunc
					addCb = func(uintptr) bool {
						defer unrefCallback(&addCb)
						if err := s.addItem(data); err != nil {
							log.Error(`Failed adding systray item`, `module`, style.SystrayID, `err`, err)
						}

						return false
					}
					glib.IdleAdd(&addCb, 0)

				case eventv1.EventKind_EVENT_KIND_DBUS_UNREGISTERSTATUSNOTIFIER:
					data, err := eventv1.DataString(evt.Data)
					if err != nil {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					log.Debug(`Deleting item`, `module`, style.SystrayID, `busName`, data)

					var deleteCb glib.SourceFunc
					deleteCb = func(uintptr) bool {
						defer unrefCallback(&deleteCb)
						if err := s.deleteItem(data); err != nil {
							log.Debug(`Failed deleting item`, `module`, style.SystrayID, `err`, err)
							return false
						}
						return false
					}
					glib.IdleAdd(&deleteCb, 0)

				case eventv1.EventKind_EVENT_KIND_DBUS_UPDATETITLE:
					data := &eventv1.UpdateTitleValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}

					var updateCb glib.SourceFunc
					updateCb = func(uintptr) bool {
						defer unrefCallback(&updateCb)
						item, ok := s.items[data.BusName]
						if !ok {
							return false
						}
						item.data.Title = data.Title
						item.updateTooltip()

						return false
					}
					glib.IdleAdd(&updateCb, 0)

				case eventv1.EventKind_EVENT_KIND_DBUS_UPDATETOOLTIP:
					data := &eventv1.UpdateTooltipValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}

					var updateCb glib.SourceFunc
					updateCb = func(uintptr) bool {
						defer unrefCallback(&updateCb)
						item, ok := s.items[data.BusName]
						if !ok {
							return false
						}
						item.data.Tooltip = data.Tooltip
						item.updateTooltip()

						return false
					}
					glib.IdleAdd(&updateCb, 0)

				case eventv1.EventKind_EVENT_KIND_DBUS_UPDATEICON:
					data := &eventv1.UpdateIconValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}

					var updateCb glib.SourceFunc
					updateCb = func(uintptr) bool {
						defer unrefCallback(&updateCb)
						item, ok := s.items[data.BusName]
						if !ok {
							return false
						}
						item.data.Icon = data.Icon
						if err := item.updateIcon(); err != nil {
							log.Debug(`Failed updating icon`, `module`, style.SystrayID, `busName`, item.data.BusName, `err`, err, `cbPtr`, uintptr(unsafe.Pointer(&updateCb)))
						}

						return false
					}
					glib.IdleAdd(&updateCb, 0)

				case eventv1.EventKind_EVENT_KIND_DBUS_UPDATESTATUS:
					data := &eventv1.UpdateStatusValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}

					var updateCb glib.SourceFunc
					updateCb = func(uintptr) bool {
						defer unrefCallback(&updateCb)
						item, ok := s.items[data.BusName]
						if !ok {
							return false
						}
						item.data.Status = data.Status
						item.updateStatus(s.clientContainer, s.clientHiddenContainer)

						return false
					}
					glib.IdleAdd(&updateCb, 0)

				case eventv1.EventKind_EVENT_KIND_DBUS_UPDATEMENU:
					data := &eventv1.UpdateMenuValue{}
					if !evt.Data.MessageIs(data) {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Error(`Invalid event`, `module`, style.SystrayID, `event`, evt)
						continue
					}

					var updateCb glib.SourceFunc
					updateCb = func(uintptr) bool {
						defer unrefCallback(&updateCb)
						item, ok := s.items[data.BusName]
						if !ok {
							return false
						}
						item.data.Menu = data.Menu
						if err := item.updateMenu(); err != nil {
							log.Debug(`Failed updating menu`, `module`, style.SystrayID, `busName`, item.data.BusName, `err`, err)
						}

						return false
					}
					glib.IdleAdd(&updateCb, 0)

				}
			}
		}
	}
}

func (s *systray) close(container *gtk.Box) {
	defer s.Unref()
	log.Debug(`Closing module on request`, `module`, style.SystrayID)
	container.Remove(&s.container.Widget)
}

func newSystray(panel *panel, cfg *modulev1.Systray) *systray {
	s := &systray{
		refTracker:    newRefTracker(),
		panel:         panel,
		cfg:           cfg,
		items:         make(map[string]*systrayItem),
		modules:       make([]module, 0),
		receivers:     make(map[module]chan<- *eventv1.Event),
		hidden:        make(chan struct{}),
		eventCh:       make(chan *eventv1.Event, 10),
		quitCh:        make(chan struct{}),
		hideInhibitor: newSystrayInhibitor(cfg.AutoHideDelay.AsDuration()),
	}

	s.AddRef(func() {
		close(s.quitCh)
		close(s.eventCh)
	})

	return s
}
