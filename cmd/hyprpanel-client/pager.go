package main

import (
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pdf/hyprpanel/internal/hypripc"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	"github.com/pdf/hyprpanel/style"
)

type sortedWorkspaces struct {
	workspaces   []*pagerWorkspace
	workspaceIdx map[string]int
}

func (s sortedWorkspaces) Len() int {
	return len(s.workspaces)
}
func (s sortedWorkspaces) Swap(i, j int) {
	s.workspaces[i], s.workspaces[j] = s.workspaces[j], s.workspaces[i]
	s.workspaceIdx[s.workspaces[i].name] = i
	s.workspaceIdx[s.workspaces[j].name] = j
}
func (s sortedWorkspaces) Less(i, j int) bool {
	if s.workspaces[i].id == s.workspaces[j].id {
		return s.workspaces[i].name < s.workspaces[j].name
	}
	return s.workspaces[i].id < s.workspaces[j].id
}
func (s sortedWorkspaces) getIdx(name string) int {
	return s.workspaceIdx[name]
}
func (s sortedWorkspaces) getWorkspace(idx int) *pagerWorkspace {
	return s.workspaces[idx]
}

func newSortedWorkspaces(workspaces map[string]*pagerWorkspace) sortedWorkspaces {
	s := sortedWorkspaces{
		workspaces:   make([]*pagerWorkspace, len(workspaces)),
		workspaceIdx: make(map[string]int),
	}
	i := 0
	for _, ws := range workspaces {
		s.workspaces[i] = ws
		s.workspaceIdx[ws.name] = i
		i++
	}
	sort.Sort(s)
	return s
}

type pager struct {
	*refTracker
	panel            *panel
	cfg              *modulev1.Pager
	scale            float64
	activeClient     string
	activeWorkspace  string
	workspaces       map[string]*pagerWorkspace
	workspaceIDs     map[int]*pagerWorkspace
	clientWorkspaces map[string]string
	sortedWorkspaces sortedWorkspaces
	eventCh          chan *eventv1.Event
	quitCh           chan struct{}

	container *gtk.Box
}

func (p *pager) getWorkspace(name string, id int) (*pagerWorkspace, error) {
	ws, ok := p.workspaces[name]
	if !ok {
		if ws, ok = p.workspaceIDs[id]; !ok {
			return nil, errNotFound
		}
	}
	return ws, nil
}

func (p *pager) addWorkspace(name string, pinned bool) *pagerWorkspace {
	ws := newPagerWorkspace(p, name, pinned)
	p.workspaces[name] = ws
	p.sortedWorkspaces = newSortedWorkspaces(p.workspaces)
	return ws
}

func (p *pager) setWorkspaceID(ws *pagerWorkspace, id int) {
	if ws.id != id {
		delete(p.workspaceIDs, ws.id)
		ws.id = id
		p.workspaceIDs[id] = ws
		p.sortedWorkspaces = newSortedWorkspaces(p.workspaces)
	}
}

func (p *pager) renameWorkspace(id int, name string) {
	ws, ok := p.workspaceIDs[id]
	if !ok {
		return
	}
	delete(p.workspaces, ws.name)
	ws.rename(name)
	p.workspaces[name] = ws
	p.sortedWorkspaces = newSortedWorkspaces(p.workspaces)
}

func (p *pager) deleteWorkspace(name string) error {
	ws, err := p.getWorkspace(name, 0)
	if err != nil {
		return err
	}

	if err := ws.close(p.container); err != nil {
		if errors.Is(err, errPinned) {
			ws.setActive(false, false)
			return nil
		}
		return err
	}

	delete(p.workspaces, name)
	delete(p.workspaceIDs, ws.id)

	p.sortedWorkspaces = newSortedWorkspaces(p.workspaces)

	return nil
}

func (p *pager) updateClient(client *hypripc.Client) {
	if prevWsName, ok := p.clientWorkspaces[client.Address]; ok {
		if prevWsName != client.Workspace.Name {
			if prevws, ok := p.workspaces[prevWsName]; ok {
				prevws.deleteClient(client.Address)
			}
		}
	}
	if ws, ok := p.workspaces[client.Workspace.Name]; ok {
		ws.updateClient(client)
	}
	p.clientWorkspaces[client.Address] = client.Workspace.Name
}

func (p *pager) deleteClient(addr string) {
	if wsName, ok := p.clientWorkspaces[addr]; ok {
		if ws, ok := p.workspaces[wsName]; ok {
			ws.deleteClient(addr)
		}
	}
	delete(p.clientWorkspaces, addr)
}

func (p *pager) update() error {
	spaces, err := p.panel.hypr.Workspaces()
	if err != nil {
		return err
	}

	clients, err := p.panel.hypr.Clients()
	if err != nil {
		return err
	}

	live := make(map[string]struct{})
	for _, space := range spaces {
		ws, err := p.getWorkspace(space.Name, space.ID)
		if err != nil {
			ws = p.addWorkspace(space.Name, false)
			if err := ws.build(p.container); err != nil {
				return err
			}
		}
		p.setWorkspaceID(ws, space.ID)

		for _, client := range clients {
			if client.Workspace.Name == space.Name {
				p.updateClient(&client)
			}
		}

		ws.setActive(true, p.activeWorkspace == space.Name)
		live[space.Name] = struct{}{}
	}

	for name := range p.workspaces {
		if _, ok := live[name]; ok {
			continue
		}

		if err := p.deleteWorkspace(name); err != nil {
			log.Debug(`Failed deleting workspace`, `module`, style.PagerID, `err`, err)
		}
	}

	return nil
}

func (p *pager) build(container *gtk.Box) error {
	activeClient, err := p.panel.hypr.ActiveWindow()
	if err != nil {
		return err
	}
	p.activeClient = activeClient.Address

	activeWorkspace, err := p.panel.hypr.ActiveWorkspace()
	if err != nil {
		return err
	}
	p.activeWorkspace = activeWorkspace.Name

	p.container = gtk.NewBox(p.panel.orientation, 0)
	p.AddRef(p.container.Unref)
	p.container.SetName(style.PagerID)
	p.container.AddCssClass(style.ModuleClass)

	scrollCb := func(_ gtk.EventControllerScroll, dx, dy float64) bool {
		target := `e+1`
		if p.cfg.ScrollIncludeInactive {
			if p.cfg.ActiveMonitorOnly {
				target = `r+1`
			} else {
				target = `+1`
			}
		} else if p.cfg.ActiveMonitorOnly {
			target = `m+1`
		}

		if dy < 0 {
			target = `e-1`
			if p.cfg.ScrollIncludeInactive {
				if p.cfg.ActiveMonitorOnly {
					target = `r-1`
				} else {
					target = `-1`
				}
			} else if p.cfg.ActiveMonitorOnly {
				target = `m-1`
			}
			if p.cfg.ScrollWrapWorkspaces && p.activeWorkspace != `` {
				currentIdx := p.sortedWorkspaces.getIdx(p.activeWorkspace)
				for idx := currentIdx - 1; ; idx-- {
					if idx < 0 {
						idx = p.sortedWorkspaces.Len() - 1
					}
					ws := p.sortedWorkspaces.getWorkspace(idx)
					if !p.cfg.ScrollIncludeInactive && !ws.live {
						continue
					}
					if idx == currentIdx {
						return true
					}

					// We can't use the name here because Hyprland will allocate random IDs to inactive workspaces
					//target = ws.name
					target = strconv.Itoa(ws.id)
					break
				}
			}
		} else if p.cfg.ScrollWrapWorkspaces && p.activeWorkspace != `` {
			currentIdx := p.sortedWorkspaces.getIdx(p.activeWorkspace)
			for idx := currentIdx + 1; ; idx++ {
				if idx >= p.sortedWorkspaces.Len() {
					idx = 0
				}
				ws := p.sortedWorkspaces.getWorkspace(idx)
				if !p.cfg.ScrollIncludeInactive && !ws.live {
					continue
				}
				if idx == currentIdx {
					return true
				}

				// We can't use the name here because Hyprland will allocate random IDs to inactive workspaces
				//target = ws.name
				target = strconv.Itoa(ws.id)
				break
			}
		}

		if err := p.panel.hypr.Dispatch(hypripc.DispatchWorkspace, target); err != nil {
			log.Error(`Failed dispatching workspace switch`, `module`, style.PagerID, `err`, err.Error())
			return false
		}
		return true
	}
	p.AddRef(func() {
		unrefCallback(&scrollCb)
	})

	scrollController := gtk.NewEventControllerScroll(gtk.EventControllerScrollVerticalValue | gtk.EventControllerScrollDiscreteValue)
	scrollController.ConnectScroll(&scrollCb)
	p.container.AddController(&scrollController.EventController)

	for i, name := range p.cfg.Pinned {
		ws := p.addWorkspace(name, true)
		// HACK: Set made up IDs because Hyprland doesn't support stable workspace ID to name mappings
		p.setWorkspaceID(ws, i+1)
		if err := ws.build(p.container); err != nil {
			return err
		}
	}

	container.Append(&p.container.Widget)

	if err := p.update(); err != nil {
		return err
	}

	go p.watch()

	return nil
}

func (p *pager) events() chan<- *eventv1.Event {
	return p.eventCh
}

func (p *pager) close(container *gtk.Box) {
	defer p.Unref()
	for _, ws := range p.workspaces {
		ws.close(p.container)
	}
	container.Remove(&p.container.Widget)
}

func (p *pager) watch() {
	ticker := time.NewTicker(time.Second)
	p.AddRef(ticker.Stop)

	for {
		select {
		case <-p.quitCh:
			return
		default:
			select {
			case <-p.quitCh:
				return
			case <-ticker.C:
			case evt := <-p.eventCh:
				switch evt.Kind {
				case eventv1.EventKind_EVENT_KIND_HYPR_WORKSPACE:
					name, err := eventv1.DataString(evt.Data)
					if err != nil {
						log.Warn(`Invalid event`, `module`, style.PagerID, `evt`, evt)
						continue
					}
					p.activeWorkspace = name
				case eventv1.EventKind_EVENT_KIND_HYPR_CREATEWORKSPACE:
				case eventv1.EventKind_EVENT_KIND_HYPR_CLOSEWINDOW:
					addr, err := eventv1.DataString(evt.Data)
					if err != nil {
						log.Warn(`Invalid event`, `module`, style.PagerID, `evt`, evt)
						continue
					}
					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer unrefCallback(&cb)
						p.deleteClient(addr)
						return false
					}
					glib.IdleAdd(&cb, 0)
					continue
				case eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOWV2:
					addr, err := eventv1.DataString(evt.Data)
					if err != nil {
						log.Warn(`Invalid event`, `module`, style.PagerID, `evt`, evt)
						continue
					}
					p.activeClient = addr
				case eventv1.EventKind_EVENT_KIND_HYPR_DESTROYWORKSPACE:
				case eventv1.EventKind_EVENT_KIND_HYPR_MOVEWORKSPACE:
				case eventv1.EventKind_EVENT_KIND_HYPR_RENAMEWORKSPACE:
					data := &eventv1.HyprRenameWorkspaceValue{}
					if !evt.Data.MessageIs(data) {
						log.Warn(`Invalid event`, `module`, style.PagerID, `evt`, evt)
						continue
					}
					if err := evt.Data.UnmarshalTo(data); err != nil {
						log.Warn(`Invalid event`, `module`, style.PagerID, `evt`, evt)
						continue
					}
					var cb glib.SourceFunc
					cb = func(uintptr) bool {
						defer unrefCallback(&cb)
						p.renameWorkspace(int(data.Id), data.Name)
						return false
					}
					glib.IdleAdd(&cb, 0)
					continue
				case eventv1.EventKind_EVENT_KIND_HYPR_OPENWINDOW:
				case eventv1.EventKind_EVENT_KIND_HYPR_FULLSCREEN:
				case eventv1.EventKind_EVENT_KIND_HYPR_WINDOWTITLE:
				case eventv1.EventKind_EVENT_KIND_HYPR_MOVEWINDOW:
				default:
					continue
				}
			}

			var cb glib.SourceFunc
			cb = func(uintptr) bool {
				defer unrefCallback(&cb)
				if err := p.update(); err != nil {
					log.Debug(`Failed updating`, `module`, style.PagerID, `err`, err)
					return false
				}

				return false
			}

			glib.IdleAdd(&cb, 0)
		}
	}
}

func newPager(panel *panel, cfg *modulev1.Pager) *pager {
	p := &pager{
		refTracker:       newRefTracker(),
		panel:            panel,
		cfg:              cfg,
		workspaces:       make(map[string]*pagerWorkspace),
		workspaceIDs:     make(map[int]*pagerWorkspace),
		clientWorkspaces: make(map[string]string),
		eventCh:          make(chan *eventv1.Event),
		quitCh:           make(chan struct{}),
	}
	p.AddRef(func() {
		close(p.quitCh)
		close(p.eventCh)
	})

	if panel.orientation == gtk.OrientationHorizontalValue {
		p.scale = float64(panel.cfg.Size) / float64(panel.currentMonitor.Height)
	} else {
		p.scale = float64(panel.cfg.Size) / float64(panel.currentMonitor.Width)
	}

	return p
}
