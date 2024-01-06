package dbus

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/hashicorp/go-hclog"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
)

type menuProp string

const (
	sniMenuName = `com.canonical.dbusmenu`
	sniMenuPath = dbus.ObjectPath(`/MenuBar`)

	sniMenuMethodAboutToShow = sniMenuName + `.AboutToShow`
	sniMenuMethodGetLayout   = sniMenuName + `.GetLayout`
	sniMenuMethodEvent       = sniMenuName + `.Event`

	sniMenuMemberItemPropertiesUpdated = `ItemsPropertiesUpdated`
	sniMenuMemberLayoutUpdated         = `LayoutUpdated`

	sniMenuSignalItemsPropertiesUpdated = sniMenuName + `.` + sniMenuMemberItemPropertiesUpdated
	sniMenuSignalLayoutUpdated          = sniMenuName + `.` + sniMenuMemberLayoutUpdated

	menuType            menuProp = `type`
	menuLabel           menuProp = `label`
	menuEnabled         menuProp = `enabled`
	menuVisible         menuProp = `visible`
	menuIconName        menuProp = `icon-name`
	menuIconData        menuProp = `icon-data`
	menuShortcut        menuProp = `shortcut`
	menuToggleType      menuProp = `toggle-type`
	menuToggleState     menuProp = `toggle-state`
	menuChildrenDisplay menuProp = `children-display`
	menuAccessibleDesc  menuProp = `accessible-desc`
)

type sniMenu struct {
	ID         int32
	Properties map[menuProp]dbus.Variant
	Children   []*sniMenu
	busName    string
	log        hclog.Logger
}

func (m *sniMenu) Parse() (*eventv1.StatusNotifierValue_Menu, error) {
	menu := &eventv1.StatusNotifierValue_Menu{
		Id:         m.ID,
		Properties: &eventv1.StatusNotifierValue_Menu_Properties{},
		Children:   make([]*eventv1.StatusNotifierValue_Menu, len(m.Children)),
	}

	for k, v := range m.Properties {
		switch k {
		case menuType:
			var val string
			if err := v.Store(&val); err != nil {
				return nil, err
			}

			switch val {
			case `standard`:
			case `separator`:
				menu.Properties.IsSeparator = true
			default:
				m.log.Warn(`Unhandled menu type`, `busName`, m.busName, `menuID`, m.ID, `menuType`, val)
			}
		case menuLabel:
			var val string
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			// TDOD: Handle menu accelerators
			menu.Properties.Label = val
		case menuEnabled:
			var val bool
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			menu.Properties.IsDisabled = !val
		case menuVisible:
			var val bool
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			menu.Properties.IsHidden = !val
		case menuIconName:
			var val string
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			menu.Properties.IconName = val
		case menuIconData:
			var val []byte
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			menu.Properties.IconData = val
		case menuToggleType:
			var val string
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			switch val {
			case `checkmark`:
				menu.Properties.IsCheckbox = true
			case `radio`:
				menu.Properties.IsRadio = true
			case ``:
				continue
			default:
				m.log.Warn(`Unknown menu toggle type, defaulting to checkbox`, `busName`, m.busName, `menuID`, m.ID, `toggleType`, val)
				menu.Properties.IsCheckbox = true
			}
		case menuToggleState:
			var val int32
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			menu.Properties.ToggleState = val
		case menuShortcut:
			// TODO: Handle shortcuts, maybe
		case menuAccessibleDesc:
			var val string
			if err := v.Store(&val); err != nil {
				return nil, err
			}
		case menuChildrenDisplay:
			var val string
			if err := v.Store(&val); err != nil {
				return nil, err
			}
			if val != `submenu` {
				return nil, fmt.Errorf("unknown menu children-display: '%s'", val)
			}
			menu.Properties.IsParent = true
		default:
			return nil, fmt.Errorf("unhandled menu prop: '%s'", k)
		}
	}

	for i, sub := range m.Children {
		sub.busName = m.busName
		sub.log = m.log
		subMenu, err := sub.Parse()
		if err != nil {
			return nil, err
		}
		menu.Children[i] = subMenu
	}

	return menu, nil
}

func newSNIMenu(logger hclog.Logger, busName string) *sniMenu {
	return &sniMenu{
		log:     logger,
		busName: busName,
	}
}
