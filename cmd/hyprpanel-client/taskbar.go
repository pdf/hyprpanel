package main

import (
	"errors"
	"time"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pdf/hyprpanel/internal/hypripc"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type taskbar struct {
	*refTracker
	*api
	cfg             *modulev1.Taskbar
	eventCh         chan *eventv1.Event
	quitCh          chan struct{}
	itemSize        uint32
	itemScale       float64
	activeClient    string
	activeWorkspace string
	items           map[string]*taskbarItem
	itemClasses     map[string]string
	pinned          map[string]string

	container *gtk.ScrolledWindow
	inner     *gtk.Box
}

func (t *taskbar) getItem(query string) (*taskbarItem, error) {
	item, ok := t.items[query]
	if !ok {
		return nil, errNotFound
	}

	return item, nil
}

func (t *taskbar) addItem(class string, client *hypripc.Client, pinned bool) error {
	itemCount := len(t.items) + 1
	t.updateItemScale(itemCount)

	item := newTaskbarItem(t.cfg, t.api, class, pinned, t.itemScale, t.itemSize, client)
	if err := item.build(t.inner); err != nil {
		return err
	}

	if t.cfg.GroupTasks || pinned {
		t.items[class] = item
	} else {
		t.items[client.Address] = item
	}

	return nil
}

func (t *taskbar) deleteClient(addr string) error {
	query := addr
	if t.cfg.GroupTasks {
		var ok bool
		query, ok = t.itemClasses[addr]
		if !ok {
			return errNotFound
		}
	} else {
		for c, a := range t.pinned {
			if a == addr {
				query = c
				t.pinned[c] = ``
			}
		}
	}
	item, err := t.getItem(query)
	if err != nil {
		return err
	}

	delete(t.itemClasses, addr)

	if err := item.deleteClient(addr, t.inner); err != nil {
		if errors.Is(err, errPinned) || errors.Is(err, errNotEmpty) {
			return nil
		}

		return err
	}

	if t.cfg.GroupTasks {
		delete(t.items, item.class)
	} else {
		delete(t.items, addr)
	}

	t.updateItemScale(len(t.items))

	return nil
}

func (t *taskbar) update() error {
	hyprclients, err := t.hypr.Clients()
	if err != nil {
		return err
	}

	for _, hyprclient := range hyprclients {
		if !hyprclient.Mapped || hyprclient.Hidden {
			continue
		}

		if t.cfg.ActiveWorkspaceOnly && hyprclient.Workspace.Name != t.activeWorkspace {
			if err := t.deleteClient(hyprclient.Address); err != nil {
				log.Trace(`Failed deleting client for current workspace`, `module`, style.TaskbarID, `err`, err)
			}
			continue
		}

		if t.cfg.ActiveMonitorOnly && hyprclient.Monitor != t.currentMonitor.ID {
			if err := t.deleteClient(hyprclient.Address); err != nil {
				log.Trace(`Failed deleting client for current monitor`, `module`, style.TaskbarID, `err`, err)
			}
			continue
		}

		hyprclient := hyprclient
		if hyprclient.Class == `` {
			if hyprclient.InitialClass != `` {
				hyprclient.Class = hyprclient.InitialClass
			} else {
				hyprclient.Class = hyprclient.InitialTitle
			}
		}
		if hyprclient.Class == `` {
			continue
		}

		pinnedAddr, pinned := t.pinned[hyprclient.Class]
		if pinnedAddr != `` && pinnedAddr != hyprclient.Address {
			pinned = false
		}
		query := hyprclient.Address
		if t.cfg.GroupTasks || pinned {
			if t.cfg.GroupTasks {
				// Check for clients that replace their class during runtime
				class, ok := t.itemClasses[hyprclient.Address]
				if ok && class != hyprclient.Class {
					if err := t.deleteClient(hyprclient.Address); err != nil {
						log.Trace(`Failed deleting obsolete client`, `module`, style.TaskbarID, `address`, hyprclient.Address, `prevClass`, class, `newClass`, hyprclient.Class, `err`, err)
					}
				}
			}
			query = hyprclient.Class
		}
		item, err := t.getItem(query)
		if err != nil {
			if err := t.addItem(hyprclient.Class, &hyprclient, pinned); err != nil {
				return err
			}
		} else {
			if pinned && pinnedAddr == `` {
				t.pinned[hyprclient.Class] = hyprclient.Address
			}
			if err := item.updateClient(&hyprclient, hyprclient.Address == t.activeClient, t.activeClient); err != nil {
				return err
			}
		}
		t.itemClasses[hyprclient.Address] = hyprclient.Class
	}

	return nil
}

func (t *taskbar) updateItemScale(itemCount int) {
	var targetSize int
	if t.orientation == gtk.OrientationHorizontalValue {
		targetSize = t.container.GetAllocatedWidth()
	} else {
		targetSize = t.container.GetAllocatedHeight()
	}
	if targetSize < int(t.cfg.MaxSize) {
		targetSize = int(t.cfg.MaxSize)
	}
	if targetSize == 0 {
		return
	}

	itemSize := t.itemSize
	itemScale := t.itemScale

	delta := float64(float64(targetSize) - (float64(itemSize) * float64(itemCount)))
	var scale float64
	if delta > 0 {
		scale = 1.0
	} else {
		scale = 1 + (delta / float64(itemCount) / float64(itemSize))
	}

	if itemScale != scale {
		t.itemScale = scale
		for _, item := range t.items {
			item.updateScale(t.itemScale, t.itemSize)
		}
	}
}

func (t *taskbar) build(container *gtk.Box) error {
	activeWorkspace, err := t.hypr.ActiveWorkspace()
	if err != nil {
		return err
	}
	t.activeWorkspace = activeWorkspace.Name
	activeWindow, err := t.hypr.ActiveWindow()
	if err != nil {
		return err
	}
	t.activeClient = activeWindow.Address

	// TODO: This is a hack due to currently being unable to create custom widgets
	// with puregotk. We need to detect when content or neighbour size changes would
	// trigger an overflow. Using a scrolled window allows us to do this by hijacking
	// the window's adjustment events.
	t.container = gtk.NewScrolledWindow()
	t.AddRef(t.container.Unref)
	t.container.SetName(style.TaskbarID)
	t.container.AddCssClass(style.ModuleClass)

	t.inner = gtk.NewBox(t.orientation, 0)
	t.AddRef(t.inner.Unref)
	t.inner.SetHalign(gtk.AlignStartValue)
	t.inner.SetValign(gtk.AlignStartValue)
	t.inner.SetHomogeneous(true)

	scaleCb := func(adj gtk.Adjustment) {
		t.updateItemScale(len(t.items))
	}
	t.AddRef(func() {
		unrefCallback(&scaleCb)
	})

	if t.orientation == gtk.OrientationHorizontalValue {
		t.container.SetHexpand(t.cfg.Expand)
		t.container.GetHadjustment().ConnectChanged(&scaleCb)
	} else {
		t.container.SetVexpand(t.cfg.Expand)
		t.container.GetVadjustment().ConnectChanged(&scaleCb)
	}

	t.container.SetChild(&t.inner.Widget)
	container.Append(&t.container.Widget)

	var updateCb glib.SourceFunc
	updateCb = func(uintptr) bool {
		defer unrefCallback(&updateCb)

		for _, c := range t.cfg.Pinned {
			if err := t.addItem(c, nil, true); err != nil {
				log.Warn(`Failed adding pinned task`, `module`, style.TaskbarID, `err`, err)
			}
		}

		if err := t.update(); err != nil {
			log.Warn(`Failed updating`, `module`, style.TaskbarID, `err`, err)
			return false
		}
		return false
	}
	glib.IdleAdd(&updateCb, 0)

	go t.watch()

	return nil
}

func (t *taskbar) close(container *gtk.Box) {
	container.Remove(&t.container.Widget)
	t.Unref()
}

func (t *taskbar) events() chan<- *eventv1.Event {
	return t.eventCh
}

func (t *taskbar) watch() {
	ticker := time.NewTicker(time.Second)
	t.AddRef(ticker.Stop)

	for {
		select {
		case <-t.quitCh:
			return
		default:
			select {
			case <-t.quitCh:
				return
			case <-ticker.C:
			case evt := <-t.eventCh:
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_HYPR_WORKSPACE:
					name, err := eventv1.DataString(evt.Data)
					if err != nil {
						log.Warn(`Invalid event`, `module`, style.TaskbarID, `evt`, evt)
						continue
					}
					t.activeWorkspace = name
				case eventv1.EventKind_EVENT_KIND_HYPR_CLOSEWINDOW:
					addr, err := eventv1.DataString(evt.Data)
					if err != nil {
						log.Warn(`Invalid event`, `module`, style.TaskbarID, `evt`, evt)
						continue
					}

					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer unrefCallback(&cb)
						if err := t.deleteClient(addr); err != nil {
							log.Debug(`Failed deleting client`, `module`, style.TaskbarID, `evt`, evt, `err`, err)
							return false
						}
						return false
					}

					glib.IdleAdd(&cb, 0)
					continue
				case eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOWV2:
					addr, err := eventv1.DataString(evt.Data)
					if err != nil {
						log.Warn(`Invalid event`, `module`, style.TaskbarID, `evt`, evt)
						continue
					}
					t.activeClient = string(addr)
				case eventv1.EventKind_EVENT_KIND_HYPR_MOVEWORKSPACE:
				case eventv1.EventKind_EVENT_KIND_HYPR_OPENWINDOW:
				case eventv1.EventKind_EVENT_KIND_HYPR_WINDOWTITLE:
				case eventv1.EventKind_EVENT_KIND_HYPR_MOVEWINDOW:
				default:
					continue
				}

				var cb glib.SourceFunc
				cb = func(uintptr) bool {
					defer unrefCallback(&cb)
					if err := t.update(); err != nil {
						log.Debug(`Failed updating`, `module`, style.TaskbarID, `err`, err)
					}
					return false
				}

				glib.IdleAdd(&cb, 0)
			}
		}
	}
}

func newTaskbar(cfg *modulev1.Taskbar, a *api) *taskbar {
	if cfg.PreviewWidth == 0 {
		cfg.PreviewWidth = 256
	}

	t := &taskbar{
		refTracker:  newRefTracker(),
		api:         a,
		cfg:         cfg,
		itemSize:    a.panelCfg.Size,
		itemScale:   1.0,
		eventCh:     make(chan *eventv1.Event),
		quitCh:      make(chan struct{}),
		items:       make(map[string]*taskbarItem),
		itemClasses: make(map[string]string),
		pinned:      make(map[string]string),
	}
	t.AddRef(func() {
		close(t.quitCh)
		close(t.eventCh)
	})

	for _, c := range t.cfg.Pinned {
		t.pinned[c] = ``
	}

	return t
}
