package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/pdf/hyprpanel/style"
)

var (
	errInvalidPixbufArray = errors.New(`invalid pixbuf array`)
)

const systrayActionNamespace = `dbusmenu`

type systrayItem struct {
	*refTracker
	tray   *systray
	data   *eventv1.StatusNotifierValue
	pinned bool
	hidden bool
	quitCh chan struct{}

	menuRefs    *refTracker
	menu        *gtk.PopoverMenu
	icon        *gtk.Image
	revealer    *gtk.Revealer
	inner       *gtk.CenterBox
	overlay     *gtk.Overlay
	container   *gtk.Box
	wrapper     *gtk.FlowBoxChild
	actionGroup *gio.SimpleActionGroup
}

func (i *systrayItem) updateIcon() error {
	if i.icon != nil {
		i.inner.SetCenterWidget(&gtk.Widget{})
		i.icon.Unref()
		i.icon = nil
	}

	var err error

	if i.data.Icon == nil {
		return errors.New(`icon missing`)
	}

	if i.data.Icon.IconName != `` {
		i.icon, err = createIcon(i.data.Icon.IconName, int(i.tray.cfg.IconSize), false, nil, i.data.Icon.IconThemePath)
		if err != nil {
			log.Warn(`Failed creating icon from theme`, `module`, style.SystrayID, `iconName`, i.data.Icon.IconName, `err`, err)
		}
	}

	if i.icon == nil && i.data.Icon.IconPixmap != nil {
		pixbuf, err := pixbufFromSNIData(i.data.Icon.IconPixmap, int(i.tray.cfg.IconSize))
		if err != nil {
			return fmt.Errorf("failed converting pixbuf data: %w", err)
		}
		i.icon = gtk.NewImageFromPixbuf(pixbuf)
		i.icon.SetPixelSize(int(i.tray.cfg.IconSize))
	}

	if i.icon == nil {
		return errors.New(`could not create icon for tray item`)
	}

	i.inner.SetCenterWidget(&i.icon.Widget)

	return nil
}

func (i *systrayItem) updateTooltip() {
	if i.data.Tooltip == nil {
		i.inner.SetTooltipText(i.data.Title)
		return
	}
	var tooltip string
	switch {
	case i.data.Tooltip.Title != `` && i.data.Tooltip.Body != ``:
		tooltip = fmt.Sprintf("%s - %s", i.data.Tooltip.Title, i.data.Tooltip.Body)
	case i.data.Tooltip.Title != ``:
		tooltip = i.data.Tooltip.Title
	case i.data.Tooltip.Body != ``:
		tooltip = i.data.Tooltip.Body
	}

	if tooltip != `` {
		i.inner.SetTooltipText(tooltip)
	}
}

func (i *systrayItem) updateMenu() error {
	if i.menu == nil {
		return nil
	}
	if i.actionGroup != nil {
		i.container.InsertActionGroup(systrayActionNamespace, &gio.SimpleActionGroup{})
		actionGroup := i.actionGroup
		defer func() {
			actionGroup.Unref()
		}()
		i.actionGroup = gio.NewSimpleActionGroup()
		i.container.InsertActionGroup(systrayActionNamespace, i.actionGroup)
	}
	if i.menuRefs != nil {
		refs := i.menuRefs
		defer func() {
			refs.Unref()
		}()
		i.menuRefs = newRefTracker()
	}

	menuXML, err := i.buildMenuXML()
	if err != nil {
		return err
	}

	builder := gtk.NewBuilderFromString(string(menuXML), len(menuXML))
	defer builder.Unref()

	menuObj := builder.GetObject(i.data.BusName)
	if menuObj == nil {
		return errors.New(`could not build menu`)
	}
	defer menuObj.Unref()

	menuModel := &gio.MenuModel{}
	menuObj.Cast(menuModel)
	i.menu.SetMenuModel(menuModel)

	return nil
}

func (i *systrayItem) buildMenuXML() ([]byte, error) {
	if !i.data.Menu.Properties.IsParent {
		log.Debug(`Invalid menu struct, top-level menu not tagged with "children-display"`)
		i.data.Menu.Properties.IsParent = true
	}

	x := menuXMLInterface{}
	section := &menuXMLMenuSection{}
	x.Menu = &menuXMLMenu{
		ID:       i.data.BusName,
		Sections: []*menuXMLMenuSection{section},
	}

	if err := i.buildMenuXMLSection(i.data.Menu, i.data.Menu.Id, x.Menu, nil, section); err != nil {
		return nil, err
	}

	b, err := xml.Marshal(x)
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), b...), err
}

func (i *systrayItem) buildMenuXMLSection(menuData *eventv1.StatusNotifierValue_Menu, parentID int32, menu *menuXMLMenu, submenu *menuXMLMenuSubmenu, section *menuXMLMenuSection) error {
	switch {
	case menuData.Properties.IsParent:
		for _, child := range menuData.Children {
			if child.Properties.IsHidden {
				continue
			}

			switch {
			case child.Properties.IsSeparator:
				section = &menuXMLMenuSection{ID: strconv.Itoa(int(child.Id))}
				parentID = child.Id
				if child.Properties.IconName != `` {
					section.Attributes = append(section.Attributes, &menuXMLAttribute{
						Name:  `icon`,
						Value: child.Properties.IconName,
					})
				}
				if child.Properties.Label != `` {
					section.Attributes = append(section.Attributes, &menuXMLAttribute{
						Name:  `label`,
						Value: child.Properties.Label,
					})
				}
				if menu != nil {
					menu.Sections = append(menu.Sections, section)
				} else if submenu != nil {
					submenu.Sections = append(submenu.Sections, section)
				}
				continue
			case child.Properties.IsParent:
				subSection := &menuXMLMenuSection{}
				submenu = &menuXMLMenuSubmenu{
					ID:       strconv.Itoa(int(child.Id)),
					Sections: []*menuXMLMenuSection{subSection},
				}
				if child.Properties.IconName != `` {
					submenu.Attributes = append(section.Attributes, &menuXMLAttribute{
						Name:  `icon`,
						Value: child.Properties.IconName,
					})
				}
				if child.Properties.Label != `` {
					submenu.Attributes = append(section.Attributes, &menuXMLAttribute{
						Name:  `label`,
						Value: child.Properties.Label,
					})
				}
				section.Submenus = append(section.Submenus, submenu)
				if err := i.buildMenuXMLSection(child, child.Id, nil, submenu, subSection); err != nil {
					return err
				}
			default:
				if err := i.buildMenuXMLSection(child, parentID, menu, submenu, section); err != nil {
					return err
				}
			}
		}
	default:
		xmlItem := &menuXMLItem{}
		if menuData.Properties.IconName != `` {
			xmlItem.Attributes = append(xmlItem.Attributes, &menuXMLAttribute{
				Name:  `icon`,
				Value: menuData.Properties.IconName,
			})
		}
		if menuData.Properties.Label != `` {
			xmlItem.Attributes = append(xmlItem.Attributes, &menuXMLAttribute{
				Name:  `label`,
				Value: menuData.Properties.Label,
			})
		}

		var action *gio.SimpleAction
		var actionName string
		switch {
		case menuData.Properties.IsCheckbox:
			actionName = fmt.Sprintf("checkbox-%d", menuData.Id)
			stateBool := false
			if menuData.Properties.ToggleState == 1 {
				stateBool = true
			}
			state := glib.NewVariantBoolean(stateBool)
			action = gio.NewSimpleActionStateful(actionName, nil, state)
		case menuData.Properties.IsRadio:
			actionMember := fmt.Sprintf("%d", menuData.Id)
			actionName = fmt.Sprintf("radio%d::%s", parentID, actionMember)
			var stateStr string
			if menuData.Properties.ToggleState == 1 {
				stateStr = actionMember
			} else {
				stateStr = strconv.Itoa(int(menuData.Properties.ToggleState))
			}
			state := glib.NewVariantString(stateStr)
			action = gio.NewSimpleActionStateful(actionName, nil, state)
			xmlItem.Attributes = append(xmlItem.Attributes, &menuXMLAttribute{
				Name:  `target`,
				Value: actionMember,
			})
		default:
			actionName = fmt.Sprintf("default-%d", menuData.Id)
			action = gio.NewSimpleAction(actionName, nil)
		}

		cb := func(action gio.SimpleAction, param uintptr) {
			i.tray.panel.host.SystrayMenuEvent(i.data.BusName, menuData.Id, hyprpanelv1.SystrayMenuEvent_SYSTRAY_MENU_EVENT_CLICKED, nil, time.Now())
		}
		i.menuRefs.AddRef(func() {
			glib.UnrefCallback(&cb)
		})

		action.SetEnabled(!menuData.Properties.IsDisabled)
		action.ConnectActivate(&cb)
		i.actionGroup.AddAction(action)
		xmlItem.Attributes = append(xmlItem.Attributes, &menuXMLAttribute{
			Name:  `action`,
			Value: systrayActionNamespace + `.` + actionName,
		})

		/*
			// Custom widgets would be required for menu icons because Gnome developers are anti-user.
				widgetID := strconv.Itoa(m.ID)
				item.Attribute = append(item.Attribute, &trayMenuXMLAttribute{
					Name:  `custom`,
					Value: widgetID,
				})

				btn := gtk.new
				container := gtk.NewBox(gtk.OrientationHorizontalValue, 0)
				if m.data.IconName != `` {

				}
				widgets[widgetID] =
		*/

		section.Items = append(section.Items, xmlItem)
	}

	return nil
}

func (i *systrayItem) buildMenu() error {
	i.menu = gtk.NewPopoverMenuFromModel(&gio.NewMenu().MenuModel)
	i.menu.SetName(i.data.BusName)
	if i.tray.cfg.AutoHideDelay.AsDuration() != 0 {
		hideInhibController := i.tray.hideInhibitor.newController()
		i.menu.AddController(&hideInhibController.EventController)
	}

	i.actionGroup = gio.NewSimpleActionGroup()

	if err := i.updateMenu(); err != nil {
		i.menu.Unref()
		i.menu = nil
		return err
	}

	i.container.Append(&i.menu.Widget)
	i.container.InsertActionGroup(systrayActionNamespace, i.actionGroup)

	/*
		For some reasion gtk_popover_set_position causes the following assertion:

		gtk_widget_get_parent: assertion 'GTK_IS_WIDGET (widget)' failed

		I've no idea why, and it doesn't seem to matter when this is called,
		but it seems to operate as expected, so leaving as is for now.
	*/
	switch i.tray.panel.cfg.Edge {
	case configv1.Edge_EDGE_TOP:
		i.menu.SetPosition(gtk.PosBottomValue)
	case configv1.Edge_EDGE_RIGHT:
		i.menu.SetPosition(gtk.PosLeftValue)
	case configv1.Edge_EDGE_BOTTOM:
		i.menu.SetPosition(gtk.PosTopValue)
	case configv1.Edge_EDGE_LEFT:
		i.menu.SetPosition(gtk.PosRightValue)
	}

	return nil
}

func (i *systrayItem) close(container *gtk.FlowBox) {
	close(i.quitCh)
	container.Remove(&i.wrapper.Widget)
}

func (i *systrayItem) autoHide(container *gtk.FlowBox, hiddenContainer *gtk.FlowBox) {
	time.AfterFunc(i.tray.cfg.AutoHideDelay.AsDuration(), func() {
		var moveCb glib.SourceFunc
		moveCb = func(uintptr) bool {
			defer glib.UnrefCallback(&moveCb)
			select {
			case <-i.quitCh:
				return false
			default:
				i.setHidden(true, container, hiddenContainer)
			}
			return false
		}

		glib.IdleAdd(&moveCb, 0)
	})
}

func (i *systrayItem) updateStatus(container *gtk.FlowBox, hiddenContainer *gtk.FlowBox) {
	if !i.pinned {
		for _, autoHide := range i.tray.cfg.AutoHideStatuses {
			if i.data.Status == autoHide {
				if i.data.Status != modulev1.Systray_STATUS_PASSIVE {
					i.setHidden(false, container, hiddenContainer)
				}
				i.autoHide(container, hiddenContainer)
				return
			}
		}
	}

	i.setHidden(false, container, hiddenContainer)

}

func (i *systrayItem) setHidden(hidden bool, container *gtk.FlowBox, hiddenContainer *gtk.FlowBox) {
	if i.hidden == hidden {
		return
	}

	var (
		initialHandle  uint32 = 0
		revealCb       func()
		revealCbHandle = &initialHandle
	)
	revealCb = func() {
		i.revealer.DisconnectSignal(*revealCbHandle)
		defer glib.UnrefCallback(&revealCb)

		if hidden {
			container.Remove(&i.wrapper.Widget)
			hiddenContainer.Append(&i.wrapper.Widget)
		} else {
			hiddenContainer.Remove(&i.wrapper.Widget)
			container.Append(&i.wrapper.Widget)
		}
		i.revealer.SetRevealChild(true)
	}
	*revealCbHandle = i.revealer.ConnectSignal(`notify::child-revealed`, &revealCb)
	i.revealer.SetRevealChild(false)

	i.hidden = hidden
	i.tray.updateRevealBtn()
}

func (i *systrayItem) build(container *gtk.FlowBox, hiddenContainer *gtk.FlowBox) error {
	i.wrapper = gtk.NewFlowBoxChild()
	i.AddRef(i.wrapper.Unref)
	i.wrapper.SetHalign(gtk.AlignCenterValue)
	i.wrapper.SetValign(gtk.AlignCenterValue)
	i.wrapper.SetCanFocus(false)
	i.wrapper.SetFocusOnClick(false)
	i.revealer = gtk.NewRevealer()
	i.AddRef(i.revealer.Unref)
	i.revealer.SetRevealChild(false)
	if i.tray.panel.orientation == gtk.OrientationHorizontalValue {
		i.revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideLeftValue)
	} else {
		i.revealer.SetTransitionType(gtk.RevealerTransitionTypeSlideUpValue)
	}

	i.container = gtk.NewBox(gtk.OrientationHorizontalValue, 0)
	i.AddRef(i.container.Unref)
	i.overlay = gtk.NewOverlay()
	i.AddRef(i.overlay.Unref)
	i.inner = gtk.NewCenterBox()
	i.AddRef(i.inner.Unref)
	i.inner.SetMarginTop(1)
	i.inner.SetMarginBottom(1)
	i.inner.SetMarginStart(1)
	i.inner.SetMarginEnd(1)

	clickController := gtk.NewGestureClick()
	clickController.SetButton(0)

	clickCb := func(ctrl gtk.GestureClick, nPress int, x, y float64) {
		ctrl.SetState(gtk.EventSequenceClaimedValue)

		button := ctrl.GetCurrentButton()
		switch int(button) {
		case gdk.BUTTON_PRIMARY:
			if err := i.tray.panel.host.SystrayActivate(i.data.BusName, int32(x), int32(y)); err != nil {
				log.Warn(`Activate item failed`, `module`, style.SystrayID, `busName`, i.data.BusName, `err`, err)
			}
		case gdk.BUTTON_MIDDLE:
			if err := i.tray.panel.host.SystraySecondaryActivate(i.data.BusName, int32(x), int32(y)); err != nil {
				log.Warn(`SecondaryActivate item failed`, `module`, style.SystrayID, `busName`, i.data.BusName, `err`, err)
			}
		case gdk.BUTTON_SECONDARY:
			if i.menu != nil {
				if err := i.tray.panel.host.SystraySecondaryActivate(i.data.BusName, int32(x), int32(y)); err != nil {
					log.Warn(`SecondaryActivate item failed`, `module`, style.SystrayID, `busName`, i.data.BusName, `err`, err)
				}

				i.menu.Popup()
				return
			}
			if err := i.tray.panel.host.SystrayMenuContextActivate(i.data.BusName, int32(x), int32(y)); err != nil {
				log.Warn(`MenuContextActivate item failed`, `module`, style.SystrayID, `busName`, i.data.BusName, `err`, err)
			}
		default:
			log.Debug(`Unhandled button`, `module`, style.SystrayID, `busName`, i.data.BusName, `button`, button)
		}
	}
	i.AddRef(func() {
		glib.UnrefCallback(&clickCb)
	})
	clickController.ConnectReleased(&clickCb)

	scrollController := gtk.NewEventControllerScroll(gtk.EventControllerScrollBothAxesValue)
	scrollCb := func(ctrl gtk.EventControllerScroll, dx, dy float64) bool {
		if dy != 0 {
			if err := i.tray.panel.host.SystrayScroll(i.data.BusName, int32(dy), hyprpanelv1.SystrayScrollOrientation_SYSTRAY_SCROLL_ORIENTATION_VERTICAL); err != nil {
				log.Warn(`Scroll item failed`, `module`, style.SystrayID, `busName`, i.data.BusName, `err`, err)
			}
		}
		if dx != 0 {
			if err := i.tray.panel.host.SystrayScroll(i.data.BusName, int32(dx), hyprpanelv1.SystrayScrollOrientation_SYSTRAY_SCROLL_ORIENTATION_HORIZONTAL); err != nil {
				log.Warn(`Scroll item failed`, `module`, style.SystrayID, `busName`, i.data.BusName, `err`, err)
			}
		}

		return true
	}
	i.AddRef(func() {
		glib.UnrefCallback(&scrollCb)
	})
	scrollController.ConnectScroll(&scrollCb)

	i.overlay.AddController(&clickController.EventController)
	i.overlay.AddController(&scrollController.EventController)

	i.container.Append(&i.overlay.Widget)
	i.overlay.SetChild(&i.inner.Widget)
	i.revealer.SetChild(&i.container.Widget)
	i.wrapper.SetChild(&i.revealer.Widget)

	if err := i.updateIcon(); err != nil {
		return err
	}

	i.updateTooltip()

	if i.data.Menu != nil {
		if err := i.buildMenu(); err != nil {
			return err
		}
	}

	container.Append(&i.wrapper.Widget)
	i.revealer.SetRevealChild(true)

	i.updateStatus(container, hiddenContainer)

	return nil
}

// Override the embedded ref tracker to manage some items that we have to manually nil during operation
func (i *systrayItem) Unref() {
	if i.icon != nil {
		i.icon.Unref()
	}
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

func newSystrayItem(tray *systray, data *eventv1.StatusNotifierValue) *systrayItem {
	persistent := false
	for _, name := range tray.cfg.Pinned {
		if data.Id == name {
			persistent = true
			break
		}
	}
	return &systrayItem{
		refTracker: newRefTracker(),
		tray:       tray,
		data:       data,
		pinned:     persistent,
		quitCh:     make(chan struct{}),
		menuRefs:   newRefTracker(),
	}
}
