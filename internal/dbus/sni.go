package dbus

import (
	"errors"

	"github.com/godbus/dbus/v5"
	"github.com/hashicorp/go-hclog"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
)

const (
	sniName = `org.kde.StatusNotifierItem`
	sniPath = dbus.ObjectPath(`/StatusNotifierItem`)

	sniPropertyID                  = sniName + `.Id`
	sniPropertyTooltip             = sniName + `.Tooltip`
	sniPropertyTitle               = sniName + `.Title`
	sniPropertyIconThemePath       = sniName + `.IconThemePath`
	sniPropertyStatus              = sniName + `.Status`
	sniPropertyWindowID            = sniName + `.WindowId`
	sniPropertyIconName            = sniName + `.IconName`
	sniPropertyIconPixmap          = sniName + `.IconPixmap`
	sniPropertyAttentionIconName   = sniName + `.AttentionIconName`
	sniPropertyAttentionIconPixmap = sniName + `.AttentionIconPixmap`
	sniPropertyMenu                = sniName + `.Menu`

	sniMethodContextMenu       = sniName + `.ContextMenu`
	sniMethodActivate          = sniName + `.Activate`
	sniMethodSecondaryActivate = sniName + `.SecondaryActivate`
	sniMethodScroll            = sniName + `.Scroll`

	sniMemberNewTitle         = `NewTitle`
	sniMemberNewIcon          = `NewIcon`
	sniMemberNewAttentionIcon = `NewAttentionIcon`
	sniMemberNewOverlayIcon   = `NewOverlayIcon`
	sniMemberNewToolTip       = `NewToolTip`
	sniMemberNewStatus        = `NewStatus`

	sniSignalNewTitle         = sniName + `.` + sniMemberNewTitle
	sniSignalNewIcon          = sniName + `.` + sniMemberNewIcon
	sniSignalNewAttentionIcon = sniName + `.` + sniMemberNewAttentionIcon
	sniSignalNewOverlayIcon   = sniName + `.` + sniMemberNewOverlayIcon
	sniSignalNewToolTip       = sniName + `.` + sniMemberNewToolTip
	sniSignalNewStatus        = sniName + `.` + sniMemberNewStatus
)

type statusNotifierItem struct {
	conn       *dbus.Conn
	log        hclog.Logger
	busName    string
	objectPath dbus.ObjectPath
	busObj     dbus.BusObject
	menuObj    dbus.BusObject
	target     *eventv1.StatusNotifierValue
}

func (i *statusNotifierItem) updateID() error {
	idProp, err := i.busObj.GetProperty(sniPropertyID)
	if err != nil {
		return err
	}
	if err := idProp.Store(&i.target.Id); err != nil {
		return err
	}
	if i.target.Id == `` {
		return errors.New(`id not found`)
	}

	return nil
}

func (i *statusNotifierItem) updateTitle() error {
	titleProp, err := i.busObj.GetProperty(sniPropertyTitle)
	if err != nil {
		return err
	}
	if err := titleProp.Store(&i.target.Title); err != nil {
		return err
	}
	if i.target.Title == `` {
		return errors.New(`title not found`)
	}

	return nil
}

func (i *statusNotifierItem) updateStatus() error {
	statusProp, err := i.busObj.GetProperty(sniPropertyStatus)
	if err != nil {
		return err
	}
	var status string
	if err := statusProp.Store(&status); err != nil {
		return err
	}

	switch status {
	case `Passive`:
		i.target.Status = modulev1.Systray_STATUS_PASSIVE
	case `Active`:
		i.target.Status = modulev1.Systray_STATUS_ACTIVE
	case `NeedsAttention`:
		i.target.Status = modulev1.Systray_STATUS_NEEDS_ATTENTION
	default:
		i.target.Status = modulev1.Systray_STATUS_UNSPECIFIED
	}

	if i.target.Status == modulev1.Systray_STATUS_UNSPECIFIED {
		return errors.New(`status not found`)
	}

	return nil
}

func (i *statusNotifierItem) updateTooltip() error {
	tooltipProp, err := i.busObj.GetProperty(sniPropertyTooltip)
	if err != nil {
		i.target.Tooltip = &eventv1.StatusNotifierValue_Tooltip{
			Title: i.target.Title,
		}

		return nil
	}
	if err := tooltipProp.Store(i.target.Tooltip); err != nil {
		return err
	}

	return nil
}

func (i *statusNotifierItem) updateIcon() error {
	if i.target.Icon == nil {
		i.target.Icon = &eventv1.StatusNotifierValue_Icon{}
	}

	var themePath string
	themePathProp, err := i.busObj.GetProperty(sniPropertyIconThemePath)
	if err != nil {
		i.log.Debug(`Missing IconThemePath from StatusNotifierItem, continuing`, `busName`, i.busName, `err`, err)
	} else if err := themePathProp.Store(&themePath); err != nil {
		return err
	}
	i.target.Icon.IconThemePath = themePath

	var status string
	statusProp, err := i.busObj.GetProperty(sniPropertyStatus)
	if err != nil {
		i.log.Debug(`Missing Status from StatusNotifierItem, continuing`, `busName`, i.busName, `err`, err)
	} else if err := statusProp.Store(&status); err != nil {
		return err
	}

	var name string
	if status == `NeedsAttention` {
		iconProp, err := i.busObj.GetProperty(sniPropertyAttentionIconName)
		if err != nil {
			i.log.Debug(`Missing AttentionIconName from StatusNotifierItem, continuing`, `busName`, i.busName, `err`, err)
		} else if err := iconProp.Store(&name); err != nil {
			return err
		}
	} else {
		iconProp, err := i.busObj.GetProperty(sniPropertyIconName)
		if err != nil {
			i.log.Debug(`Missing IconName from StatusNotifierItem, continuing`, `busName`, i.busName, `err`, err)
		} else if err := iconProp.Store(&name); err != nil {
			return err
		}
	}
	i.target.Icon.IconName = name

	pixbufArr := make([]*eventv1.StatusNotifierValue_Pixmap, 0)
	if status == `NeedsAttention` {
		pixbufArrProp, err := i.busObj.GetProperty(sniPropertyAttentionIconPixmap)
		if err == nil {
			if err := pixbufArrProp.Store(&pixbufArr); err != nil {
				return err
			}
		}
	} else {
		pixbufArrProp, err := i.busObj.GetProperty(sniPropertyIconPixmap)
		if err == nil {
			if err := pixbufArrProp.Store(&pixbufArr); err != nil {
				return err
			}
		}
	}

	if len(pixbufArr) == 0 && i.target.Icon.IconName == `` {
		return errors.New(`icon not found`)
	}
	i.target.Icon.IconPixmap = nil
	for _, buf := range pixbufArr {
		if i.target.Icon.IconPixmap == nil {
			i.target.Icon.IconPixmap = &eventv1.StatusNotifierValue_Pixmap{
				Width:  int32(buf.Width),
				Height: int32(buf.Height),
				Data:   buf.Data,
			}
			continue
		}
		if buf.Width > i.target.Icon.IconPixmap.Height || buf.Height > i.target.Icon.IconPixmap.Height {
			i.target.Icon.IconPixmap = &eventv1.StatusNotifierValue_Pixmap{
				Width:  int32(buf.Width),
				Height: int32(buf.Height),
				Data:   buf.Data,
			}
		}
	}

	return nil
}

func (i *statusNotifierItem) getMenuPath() error {
	menuPathProp, err := i.busObj.GetProperty(sniPropertyMenu)
	if err != nil {
		return err
	}

	var menuPath string
	if err := menuPathProp.Store(&menuPath); err != nil {
		return err
	}
	if menuPath == `` {
		return errors.New(`menu empty`)
	}

	i.menuObj = i.conn.Object(i.busName, dbus.ObjectPath(menuPath))
	if i.menuObj == nil {
		return errors.New(`menu path provided but invalid`)
	}

	return nil
}

func (i *statusNotifierItem) updateMenu() error {
	if i.menuObj == nil {
		return errUnsupported
	}

	dmenu := newSNIMenu(i.log, i.target.BusName)
	layoutCall := i.menuObj.Call(sniMenuMethodGetLayout, 0, 0, -1, []string{})
	if err := layoutCall.Store(&i.target.MenuRevision, dmenu); err != nil {
		return err
	}

	var err error
	i.target.Menu, err = dmenu.Parse()
	if err != nil {
		return err
	}

	return nil
}

func newStatusNotifierItem(conn *dbus.Conn, logger hclog.Logger, busName string, objectPath dbus.ObjectPath, busObj dbus.BusObject) (*statusNotifierItem, error) {
	i := &statusNotifierItem{
		conn:       conn,
		log:        logger,
		busName:    busName,
		objectPath: objectPath,
		busObj:     busObj,
		target: &eventv1.StatusNotifierValue{
			BusName:    busName,
			ObjectPath: string(objectPath),
		},
	}

	if err := i.updateID(); err != nil {
		return nil, err
	}

	if err := i.updateStatus(); err != nil {
		i.log.Debug(`SNI item update status failed`, `busName`, i.busName, `err`, err)
	}

	if err := i.updateIcon(); err != nil {
		return nil, err
	}

	if err := i.updateTitle(); err != nil {
		i.log.Debug(`SNI item update title failed`, `busName`, busName, `err`, err)
	}

	if err := i.updateTooltip(); err != nil {
		i.log.Debug(`SNI item update tooltip failed`, `busName`, busName, `err`, err)
	}

	if err := i.getMenuPath(); err != nil {
		i.log.Debug(`SNI item get menu path failed`, `busName`, busName, `err`, err)
	}

	if err := i.updateMenu(); err != nil {
		i.log.Debug(`SNI item update menu failed`, `busName`, busName, `err`, err)
	}

	return i, nil
}
