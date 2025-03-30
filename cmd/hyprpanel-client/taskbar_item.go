package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pdf/hyprpanel/internal/hypripc"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/pdf/hyprpanel/style"
)

const taskbarActionNamespace = `taskbarmenu`

var (
	errPinned   = errors.New(`pinned`)
	errNotEmpty = errors.New(`not empty`)
)

type sortableClients []*hypripc.Client

func (s sortableClients) Len() int {
	return len(s)
}

func (s sortableClients) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortableClients) Less(i, j int) bool {
	return s[i].Address < s[j].Address
}

type taskbarItem struct {
	*refTracker
	*api
	cfg             *modulev1.Taskbar
	class           string
	pinned          bool
	scale           float64
	size            uint32
	activeClient    string
	activeClientIdx int
	title           string
	appInfo         *hyprpanelv1.AppInfo
	clients         map[string]*hypripc.Client
	sortedClients   sortableClients
	menuRefs        *refTracker

	wrapper       *gtk.Overlay
	iconContainer *gtk.CenterBox
	container     *gtk.Box
	icon          *gtk.Image
	actionGroup   *gio.SimpleActionGroup
	menu          *gtk.PopoverMenu
	indicator     *gtk.Box
}

func (i *taskbarItem) updateMenu() error {
	if i.menu == nil {
		return nil
	}
	if i.actionGroup != nil {
		i.wrapper.InsertActionGroup(taskbarActionNamespace, &gio.SimpleActionGroup{})
		actionGroup := i.actionGroup
		defer actionGroup.Unref()
		i.actionGroup = gio.NewSimpleActionGroup()
		i.wrapper.InsertActionGroup(taskbarActionNamespace, i.actionGroup)
	}
	if i.menuRefs != nil {
		refs := i.menuRefs
		defer func() {
			refs.Unref()
		}()
		i.menuRefs = newRefTracker()
	}

	id, menuXML, err := i.buildMenuXML()
	if err != nil {
		return err
	}

	builder := gtk.NewBuilderFromString(string(menuXML), len(menuXML))
	defer builder.Unref()

	menuObj := builder.GetObject(id)
	if menuObj == nil {
		return errors.New(`could not build menu`)
	}
	defer menuObj.Unref()

	menuModel := &gio.MenuModel{}
	menuObj.Cast(menuModel)
	i.menu.SetMenuModel(menuModel)

	return nil
}

func (i *taskbarItem) buildMenu() error {
	i.actionGroup = gio.NewSimpleActionGroup()
	id, menuXML, err := i.buildMenuXML()
	if err != nil {
		return err
	}

	builder := gtk.NewBuilderFromString(string(menuXML), len(menuXML))
	defer builder.Unref()
	menuObj := builder.GetObject(id)
	defer menuObj.Unref()
	if menuObj != nil {
		menuModel := &gio.MenuModel{}
		menuObj.Cast(menuModel)
		i.menu = gtk.NewPopoverMenuFromModel(menuModel)
		switch i.panelCfg.Edge {
		case configv1.Edge_EDGE_TOP:
			i.menu.SetPosition(gtk.PosBottomValue)
		case configv1.Edge_EDGE_RIGHT:
			i.menu.SetPosition(gtk.PosLeftValue)
		case configv1.Edge_EDGE_BOTTOM:
			i.menu.SetPosition(gtk.PosTopValue)
		case configv1.Edge_EDGE_LEFT:
			i.menu.SetPosition(gtk.PosRightValue)
		}
	}

	i.container.Append(&i.menu.Widget)
	i.wrapper.InsertActionGroup(taskbarActionNamespace, i.actionGroup)

	return nil
}

func (i *taskbarItem) buildMenuXML() (string, []byte, error) {
	id := i.class
	if i.activeClient != `` {
		id = i.activeClient
	}
	x := menuXMLInterface{
		Menu: &menuXMLMenu{
			ID: id,
		},
	}

	if len(i.clients) > 0 {
		section := &menuXMLMenuSection{}
		for _, c := range i.sortedClients {
			c := c
			actionName := `focus-` + c.Address
			section.Items = append(section.Items, &menuXMLItem{
				Attributes: []*menuXMLAttribute{
					{
						Name:  `label`,
						Value: c.Title,
					},
					{
						Name:  `icon`,
						Value: i.appInfo.Icon,
					},
					{
						Name:  `action`,
						Value: taskbarActionNamespace + `.` + actionName,
					},
				},
			})
			actionCb := func(action gio.SimpleAction, param uintptr) {
				if err := i.hypr.Dispatch(hypripc.DispatchFocusWindow, `address:`+c.Address); err != nil {
					log.Debug(`Focus window failed`, `module`, style.TaskbarID, `err`, err)
				}
			}
			i.menuRefs.AddRef(func() {
				unrefCallback(&actionCb)
			})
			var action *gio.SimpleAction
			if i.cfg.GroupTasks && len(i.clients) > 1 {
				state := glib.NewVariantBoolean(c.Address == i.activeClient)
				action = gio.NewSimpleActionStateful(actionName, nil, state)
			} else {
				action = gio.NewSimpleAction(actionName, nil)
			}
			action.SetEnabled(true)
			action.ConnectActivate(&actionCb)
			i.actionGroup.AddAction(action)
		}

		x.Menu.Sections = append(x.Menu.Sections, section)
	} else {
		actionName := `launch-` + i.class
		section := &menuXMLMenuSection{
			Items: []*menuXMLItem{
				{
					Attributes: []*menuXMLAttribute{
						{
							Name:  `label`,
							Value: `Launch`,
						},
						{
							Name:  `icon`,
							Value: i.appInfo.Icon,
						},
						{
							Name:  `action`,
							Value: taskbarActionNamespace + `.` + actionName,
						},
					},
				},
			},
		}
		actionCb := func(action gio.SimpleAction, param uintptr) {
			i.launchIndicator()
			if err := i.host.Exec(&hyprpanelv1.AppInfo_Action{Name: i.appInfo.Name, Icon: i.appInfo.Icon, Exec: i.appInfo.Exec}); err != nil {
				log.Warn(`Failed launching application`, `module`, style.SystrayID, `cmd`, i.appInfo.Exec, `err`, err)
			}
		}
		i.menuRefs.AddRef(func() {
			unrefCallback(&actionCb)
		})
		action := gio.NewSimpleAction(actionName, nil)
		action.SetEnabled(true)
		action.ConnectActivate(&actionCb)
		i.actionGroup.AddAction(action)

		x.Menu.Sections = append(x.Menu.Sections, section)
	}

	if len(i.appInfo.Actions) > 0 {
		section := &menuXMLMenuSection{}
		for _, a := range i.appInfo.Actions {
			a := a
			actionName := `desktop-` + a.Name
			section.Items = append(section.Items, &menuXMLItem{
				Attributes: []*menuXMLAttribute{
					{
						Name:  `label`,
						Value: a.Name,
					},
					{
						Name:  `icon`,
						Value: a.Icon,
					},
					{
						Name:  `action`,
						Value: taskbarActionNamespace + `.` + actionName,
					},
				},
			})
			actionCb := func(action gio.SimpleAction, param uintptr) {
				if err := i.host.Exec(a); err != nil {
					log.Warn(`Failed launching application`, `module`, style.SystrayID, `cmd`, a.Exec, `err`, err)
				}
			}
			i.menuRefs.AddRef(func() {
				unrefCallback(&actionCb)
			})
			action := gio.NewSimpleAction(actionName, nil)
			action.SetEnabled(true)
			action.ConnectActivate(&actionCb)
			i.actionGroup.AddAction(action)
		}

		x.Menu.Sections = append(x.Menu.Sections, section)
	}

	if i.activeClient != `` {
		actionName := `close-` + id
		x.Menu.Sections = append(x.Menu.Sections, &menuXMLMenuSection{
			Items: []*menuXMLItem{
				{
					Attributes: []*menuXMLAttribute{
						{
							Name:  `label`,
							Value: `Close`,
						},
						{
							Name:  `action`,
							Value: taskbarActionNamespace + `.` + actionName,
						},
					},
				},
			},
		})
		actionCb := func(action gio.SimpleAction, param uintptr) {
			if err := i.hypr.Dispatch(hypripc.DispatchCloseWindow, `address:`+id); err != nil {
				log.Debug(`Close window failed`, `module`, style.TaskbarID, `err`, err)
			}
		}
		i.menuRefs.AddRef(func() {
			unrefCallback(&actionCb)
		})
		action := gio.NewSimpleAction(actionName, nil)
		action.SetEnabled(true)
		action.ConnectActivate(&actionCb)
		i.actionGroup.AddAction(action)
	}

	if len(x.Menu.Sections) == 0 {
		return ``, nil, errors.New(`empty menu`)
	}

	b, err := xml.Marshal(x)
	return id, b, err
}

func (i *taskbarItem) updateScale(scale float64, size uint32) {
	i.scale = scale
	i.size = size
	scaledSize := int(math.Floor(float64(i.size) * i.scale))
	if i.orientation == gtk.OrientationHorizontalValue {
		i.container.SetSizeRequest(scaledSize, int(i.size))
	} else {
		i.container.SetSizeRequest(int(i.size), scaledSize)
	}
	i.icon.SetPixelSize(int(math.Floor(float64(i.cfg.IconSize) * i.scale)))
}

func (i *taskbarItem) launchIndicator() {
	spinner := gtk.NewSpinner()
	spinner.SetSizeRequest(8, 8)
	if i.orientation == gtk.OrientationHorizontalValue {
		spinner.SetHexpand(true)
		spinner.SetHalign(gtk.AlignCenterValue)
	} else {
		spinner.SetVexpand(true)
		spinner.SetValign(gtk.AlignCenterValue)
	}
	spinner.Start()
	i.indicator.Append(&spinner.Widget)
	go func() {
		<-time.After(7 * time.Second)

		var cb glib.SourceFunc
		cb = func(uintptr) bool {
			defer unrefCallback(&cb)
			i.updateIndicator()
			return false
		}

		glib.IdleAdd(&cb, 0)
	}()
}

func (i *taskbarItem) updateIndicator() {
	if i.cfg.HideIndicators {
		return
	}
	for c := i.indicator.GetLastChild(); c != nil; c = i.indicator.GetFirstChild() {
		i.indicator.Remove(c)
		c.Unref()
	}

	for n := range i.sortedClients {
		c := gtk.NewBox(i.orientation, 0)
		defer c.Unref()
		if i.orientation == gtk.OrientationHorizontalValue {
			c.SetHexpand(true)
			c.SetSizeRequest(-1, 4)
		} else {
			c.SetVexpand(true)
			c.SetSizeRequest(4, -1)
		}

		switch i.panelCfg.Edge {
		case configv1.Edge_EDGE_TOP:
			c.SetValign(gtk.AlignEndValue)
		case configv1.Edge_EDGE_RIGHT:
			c.SetHalign(gtk.AlignStartValue)
		case configv1.Edge_EDGE_LEFT:
			c.SetHalign(gtk.AlignEndValue)
		case configv1.Edge_EDGE_BOTTOM:
			c.SetValign(gtk.AlignStartValue)
		}

		c.AddCssClass(style.IndicatorClass)
		i.indicator.Append(&c.Widget)
		if n == 5 {
			break
		}
	}
}

func (i *taskbarItem) updateClient(client *hypripc.Client, active bool, activeClient string) error {
	updated := false

	if i.activeClient == `` || (active && i.activeClient != client.Address) {
		i.activeClient = client.Address
		for n, c := range i.sortedClients {
			if c.Address == i.activeClient {
				i.activeClientIdx = n
				break
			}
		}

		updated = true
	}

	if _, ok := i.clients[client.Address]; !ok {
		i.clients[client.Address] = client
		i.sortedClients = append(i.sortedClients, client)
		sort.Sort(i.sortedClients)
		updated = true
	}
	var tooltip string
	if client, ok := i.clients[i.activeClient]; ok {
		tooltip = client.Title
		if i.activeClient == activeClient {
			if !i.container.HasCssClass(style.ActiveClass) {
				i.container.AddCssClass(style.ActiveClass)
			}
		} else if i.container.HasCssClass(style.ActiveClass) {
			i.container.RemoveCssClass(style.ActiveClass)
		}
	}
	if tooltip == `` {
		tooltip = i.appInfo.Name
	}
	if i.title != tooltip {
		i.title = tooltip
		updated = true
	}

	if updated {
		i.updateIndicator()
		return i.updateMenu()
	}

	return nil
}

func (i *taskbarItem) deleteClient(addr string, container *gtk.Box) error {
	_, ok := i.clients[addr]
	if !ok {
		return errNotFound
	}

	delete(i.clients, addr)
	for n, c := range i.sortedClients {
		if c.Address == addr {
			i.sortedClients = append(i.sortedClients[:n], i.sortedClients[n+1:]...)
			break
		}
	}
	i.updateIndicator()

	if len(i.clients) == 0 || i.activeClient == addr {
		if i.container.HasCssClass(style.ActiveClass) {
			i.container.RemoveCssClass(style.ActiveClass)
		}
		i.activeClient = ``
		i.activeClientIdx = 0
	}

	if err := i.updateMenu(); err != nil {
		return err
	}

	if len(i.clients) == 0 {
		return i.close(container)
	}

	return errNotEmpty
}

func (i *taskbarItem) clientTitle() string {
	if i.activeClient == `` {
		return i.appInfo.Name
	}
	if c, ok := i.clients[i.activeClient]; ok {
		return c.Title
	}

	return i.appInfo.Name
}

func (i *taskbarItem) clientSubtitle() string {
	c, ok := i.clients[i.activeClient]
	if !ok {
		return ``
	}
	mem, err := memKb(int(c.Pid))
	if err != nil {
		return ``
	}
	return fmt.Sprintf("Memory: %.1f MB", float64(mem)/1024.0)
}

func (i *taskbarItem) clientAddress() string {
	return i.activeClient
}

func (i *taskbarItem) shouldPreview() bool {
	return i.activeClient != ``
}

func (i *taskbarItem) build(container *gtk.Box) error {
	var err error
	i.appInfo, err = i.host.FindApplication(i.class)
	if err != nil {
		return err
	}
	icon, err := createIcon(i.appInfo.Icon, int(i.cfg.IconSize), false, nil)
	if err != nil {
		return err
	}
	i.icon = icon
	i.AddRef(i.icon.Unref)

	i.wrapper = gtk.NewOverlay()
	i.AddRef(i.wrapper.Unref)
	i.iconContainer = gtk.NewCenterBox()
	i.AddRef(i.iconContainer.Unref)
	i.container = gtk.NewBox(gtk.OrientationVerticalValue, 0)
	i.AddRef(i.container.Unref)
	i.container.AddCssClass(style.ClientClass)
	i.iconContainer.SetVexpand(true)
	i.iconContainer.SetHexpand(true)

	i.indicator = gtk.NewBox(i.orientation, 2)
	i.AddRef(i.indicator.Unref)
	i.indicator.SetMarginTop(4)
	i.indicator.SetMarginEnd(4)
	i.indicator.SetMarginBottom(4)
	i.indicator.SetMarginStart(4)

	switch i.panelCfg.Edge {
	case configv1.Edge_EDGE_TOP:
		i.indicator.SetHalign(gtk.AlignCenterValue)
		i.indicator.SetValign(gtk.AlignEndValue)
	case configv1.Edge_EDGE_RIGHT:
		i.indicator.SetHalign(gtk.AlignStartValue)
		i.indicator.SetValign(gtk.AlignCenterValue)
	case configv1.Edge_EDGE_LEFT:
		i.indicator.SetHalign(gtk.AlignEndValue)
		i.indicator.SetValign(gtk.AlignCenterValue)
	case configv1.Edge_EDGE_BOTTOM:
		i.indicator.SetHalign(gtk.AlignCenterValue)
		i.indicator.SetValign(gtk.AlignStartValue)
	}

	if i.orientation == gtk.OrientationHorizontalValue {
		i.indicator.SetSizeRequest(int(i.size)/2, 4)
		i.wrapper.SetHalign(gtk.AlignStartValue)
		i.wrapper.SetValign(gtk.AlignCenterValue)
	} else {
		i.indicator.SetSizeRequest(4, int(i.size)/2)
		i.wrapper.SetHalign(gtk.AlignCenterValue)
		i.wrapper.SetValign(gtk.AlignStartValue)
	}

	i.wrapper.AddOverlay(&i.indicator.Widget)
	i.iconContainer.SetCenterWidget(&i.icon.Widget)
	i.container.Append(&i.iconContainer.Widget)

	if err := i.buildMenu(); err != nil {
		return err
	}

	if i.activeClient != `` {
		i.updateIndicator()
		if err := i.updateClient(i.clients[i.activeClient], true, i.activeClient); err != nil {
			return err
		}
	}

	clickController := gtk.NewGestureClick()
	clickController.SetButton(0)

	clickCb := func(ctrl gtk.GestureClick, nPress int, x, y float64) {
		ctrl.SetState(gtk.EventSequenceClaimedValue)

		button := ctrl.GetCurrentButton()
		switch int(button) {
		case gdk.BUTTON_PRIMARY:
			if i.activeClient != `` {
				if err := i.hypr.Dispatch(hypripc.DispatchFocusWindow, `address:`+i.activeClient); err != nil {
					log.Warn(`Focus client failed`, `module`, style.TaskbarID, `err`, err)
				}
			} else {
				i.launchIndicator()
				if err := i.host.Exec(&hyprpanelv1.AppInfo_Action{Name: i.appInfo.Name, Icon: i.appInfo.Icon, Exec: i.appInfo.Exec}); err != nil {
					log.Warn(`Failed launching application`, `module`, style.SystrayID, `cmd`, i.appInfo.Exec, `err`, err)
				}
			}
		case gdk.BUTTON_MIDDLE:
			i.launchIndicator()
			if err := i.host.Exec(&hyprpanelv1.AppInfo_Action{Name: i.appInfo.Name, Icon: i.appInfo.Icon, Exec: i.appInfo.Exec}); err != nil {
				log.Warn(`Failed launching application`, `module`, style.SystrayID, `cmd`, i.appInfo.Exec, `err`, err)
			}
		case gdk.BUTTON_SECONDARY:
			if i.menu != nil {
				i.menu.Popup()
			}
		default:
			log.Debug(`Unhandled button`, `module`, style.TaskbarID, `button`, button)
		}
	}
	i.AddRef(func() {
		unrefCallback(&clickCb)
	})
	clickController.ConnectReleased(&clickCb)
	i.wrapper.AddController(&clickController.EventController)

	enterCb := func(ctrl gtk.EventControllerMotion, x, y float64) {
		ctrl.GetWidget().AddCssClass(style.HoverClass)
	}
	leaveCb := func(ctrl gtk.EventControllerMotion) {
		ctrl.GetWidget().RemoveCssClass(style.HoverClass)
	}

	motionController := gtk.NewEventControllerMotion()
	i.AddRef(func() {
		unrefCallback(&enterCb)
	})
	i.AddRef(func() {
		unrefCallback(&leaveCb)
	})
	motionController.ConnectEnter(&enterCb)
	motionController.ConnectLeave(&leaveCb)
	i.container.AddController(&motionController.EventController)

	if i.cfg.GroupTasks {
		scrollController := gtk.NewEventControllerScroll(gtk.EventControllerScrollVerticalValue | gtk.EventControllerScrollDiscreteValue)
		scrollCb := func(ctrl gtk.EventControllerScroll, dx, dy float64) bool {
			if len(i.sortedClients) == 0 {
				return true
			}
			idx := i.activeClientIdx
			if dy < 0 {
				if idx == 0 {
					idx = len(i.sortedClients) - 1
				} else {
					idx--
				}
			} else {
				if idx == len(i.sortedClients)-1 {
					idx = 0
				} else {
					idx++
				}
			}

			if err := i.hypr.Dispatch(hypripc.DispatchFocusWindow, `address:`+i.sortedClients[idx].Address); err != nil {
				log.Warn(`Focus client failed`, `module`, style.TaskbarID, `err`, err)
			}

			return true
		}
		scrollController.ConnectScroll(&scrollCb)
		i.wrapper.AddController(&scrollController.EventController)
	}

	i.container.SetHasTooltip(true)
	previewHeight := int(i.cfg.PreviewWidth * 9 / 16)
	tooltipCb := tooltipPreview(i, int(i.cfg.PreviewWidth), previewHeight)
	i.AddRef(func() { unrefCallback(&tooltipCb) })
	i.container.ConnectQueryTooltip(&tooltipCb)

	i.updateScale(i.scale, i.size)

	i.wrapper.SetChild(&i.container.Widget)
	container.Append(&i.wrapper.Widget)

	return nil
}

func (i *taskbarItem) close(container *gtk.Box) error {
	if i.pinned {
		return errPinned
	}
	container.Remove(&i.wrapper.Widget)
	i.Unref()

	return nil
}

// Override the embedded ref tracker to manage some items that we have to manually nil during operation
func (i *taskbarItem) Unref() {
	if i.menu != nil {
		i.menu.Unref()
	}
	if i.actionGroup != nil {
		i.actionGroup.Unref()
	}
	if i.menuRefs != nil {
		i.menuRefs.Unref()
	}
	i.refTracker.Unref()
}

func newTaskbarItem(cfg *modulev1.Taskbar, a *api, class string, pinned bool, scale float64, size uint32, client *hypripc.Client) *taskbarItem {
	i := &taskbarItem{
		refTracker:    newRefTracker(),
		api:           a,
		cfg:           cfg,
		class:         class,
		pinned:        pinned,
		scale:         scale,
		size:          size,
		clients:       make(map[string]*hypripc.Client),
		sortedClients: make(sortableClients, 0),
		menuRefs:      newRefTracker(),
	}
	if client != nil {
		i.clients[client.Address] = client
		i.activeClient = client.Address
		i.sortedClients = append(i.sortedClients, client)
	}

	return i
}
