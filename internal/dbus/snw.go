package dbus

import (
	"fmt"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
	"github.com/hashicorp/go-hclog"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	snwName = `org.kde.StatusNotifierWatcher`
	snwPath = dbus.ObjectPath(`/StatusNotifierWatcher`)

	snwSignalHostRegistered   = snwName + `.StatusNotifierHostRegistered`
	snwSignalHostUnregistered = snwName + `.StatusNotifierHostUnregistered`
	snwSignalItemRegistered   = snwName + `.StatusNotifierItemRegistered`
	snwSignalItemUnregistered = snwName + `.StatusNotifierItemUnregistered`
)

type statusNotifierWatcher struct {
	sync.RWMutex
	conn  *dbus.Conn
	log   hclog.Logger
	props *prop.Properties

	items       map[string]*statusNotifierItem
	itemSenders map[string]*statusNotifierItem
	eventCh     chan *eventv1.Event
	signals     chan *dbus.Signal
	quitCh      chan struct{}
}

func (s *statusNotifierWatcher) registerStatusNotifierItem(busName string, objectPath dbus.ObjectPath, sender dbus.Sender, busObj dbus.BusObject) error {
	item, err := newStatusNotifierItem(s.conn, s.log, busName, objectPath, busObj)
	if err != nil {
		return err
	}
	s.Lock()
	s.items[busName] = item
	s.itemSenders[string(sender)] = item
	s.Unlock()

	data, err := anypb.New(item.target)
	if err != nil {
		return err
	}

	s.log.Trace(`Sending SNI registration`, `busName`, busName, `objectPath`, objectPath, `sender`, sender)
	s.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_DBUS_REGISTERSTATUSNOTIFIER,
		Data: data,
	}

	if err := s.conn.Emit(snwPath, snwSignalItemRegistered, fmt.Sprintf("%s%s", busName, objectPath)); err != nil {
		return err
	}

	return nil
}

func (s *statusNotifierWatcher) RegisterStatusNotifierItem(target string, sender dbus.Sender) *dbus.Error {
	var (
		busName    string
		objectPath dbus.ObjectPath
	)
	if len(target) > 0 && target[0] == '/' {
		busName = string(sender)
		objectPath = dbus.ObjectPath(target)
	} else {
		busName = target
		objectPath = sniPath
	}
	s.log.Trace(`Received SNI register request`, `busName`, busName, `objectPath`, objectPath, `sender`, sender)

	busObj := s.conn.Object(busName, objectPath)
	if !busObj.Path().IsValid() {
		s.log.Error(`Returning DBUS error`, `busName`, busName, `objectPath`, `sender`, sender, objectPath, `err`, `invalid objectPath`)
		return &dbus.ErrMsgInvalidArg
	}

	go func() {
		if err := s.registerStatusNotifierItem(busName, objectPath, sender, busObj); err != nil {
			s.log.Warn(`Failed registering SNI`, `busName`, busName, `objectPath`, objectPath, `sender`, sender, `err`, err)
		}
	}()

	return nil
}

func (s *statusNotifierWatcher) RegisterStatusNotifierHost(_ string, _ dbus.Sender) *dbus.Error {
	return errDbusNotSupported
}

func (s *statusNotifierWatcher) RegisteredStatusNotifierItems() []string {
	s.RLock()
	defer s.RUnlock()
	items := make([]string, len(s.items))
	i := 0
	for busName := range s.items {
		items[i] = busName
		i++
	}

	return items
}

func (s *statusNotifierWatcher) IsStatusNotifierHostRegistered() bool {
	return true
}

func (s *statusNotifierWatcher) ProtocolVersion() int32 {
	return 0
}

func (s *statusNotifierWatcher) Activate(busName string, x, y int32) error {
	s.RLock()
	defer s.RUnlock()
	item, ok := s.items[busName]
	if !ok {
		return fmt.Errorf("systray item not found for busName '%s'", busName)
	}

	if call := item.busObj.Call(sniMethodActivate, 0, x, y); call.Err != nil {
		return call.Err
	}

	return nil
}

func (s *statusNotifierWatcher) SecondaryActivate(busName string, x, y int32) error {
	s.RLock()
	defer s.RUnlock()
	item, ok := s.items[busName]
	if !ok {
		return fmt.Errorf("systray item not found for busName '%s'", busName)
	}

	if call := item.busObj.Call(sniMethodSecondaryActivate, 0, x, y); call.Err != nil {
		return call.Err
	}

	return nil
}

func (s *statusNotifierWatcher) Scroll(busName string, delta int32, orientation hyprpanelv1.SystrayScrollOrientation) error {
	s.RLock()
	defer s.RUnlock()
	item, ok := s.items[busName]
	if !ok {
		return fmt.Errorf("systray item not found for busName '%s'", busName)
	}

	var orient string
	if orientation == hyprpanelv1.SystrayScrollOrientation_SYSTRAY_SCROLL_ORIENTATION_HORIZONTAL {
		orient = `horizontal`
	} else {
		orient = `vertical`
	}

	if call := item.busObj.Call(sniMethodScroll, 0, delta, orient); call.Err != nil {
		return call.Err
	}

	return nil
}

func (s *statusNotifierWatcher) MenuContextActivate(busName string, x, y int32) error {
	s.RLock()
	defer s.RUnlock()
	item, ok := s.items[busName]
	if !ok {
		return fmt.Errorf("systray item not found for busName '%s'", busName)
	}

	if call := item.busObj.Call(sniMethodContextMenu, 0, x, y); call.Err != nil {
		return call.Err
	}

	return nil
}

func (s *statusNotifierWatcher) MenuAboutToShow(busName string, menuItemID string) error {
	s.RLock()
	defer s.RUnlock()
	item, ok := s.items[busName]
	if !ok {
		return fmt.Errorf("systray item not found for busName '%s'", busName)
	}

	if call := item.menuObj.Call(sniMenuMethodAboutToShow, 0, menuItemID, true); call.Err != nil {
		// Hack around some implementations that accept only (i), not (ib)
		if call := item.menuObj.Call(sniMenuMethodAboutToShow, 0, menuItemID); call.Err != nil {
			return call.Err
		}
	}

	return nil
}

func (s *statusNotifierWatcher) MenuEvent(busName string, id int32, eventID hyprpanelv1.SystrayMenuEvent, _ any, timestamp time.Time) error {
	s.RLock()
	defer s.RUnlock()
	item, ok := s.items[busName]
	if !ok {
		return fmt.Errorf("systray item not found for busName '%s'", busName)
	}

	var event string
	switch eventID {
	case hyprpanelv1.SystrayMenuEvent_SYSTRAY_MENU_EVENT_CLICKED:
		event = `clicked`
	case hyprpanelv1.SystrayMenuEvent_SYSTRAY_MENU_EVENT_HOVERED:
		event = `hovered`
	default:
		return fmt.Errorf(`systray menu eventId of unknown type`)
	}

	// TODO: Implement data? Currently unused
	if call := item.menuObj.Call(sniMenuMethodEvent, 0, id, event, dbus.MakeVariant(``), uint32(timestamp.Unix())); call.Err != nil {
		return call.Err
	}

	return nil
}

// getItem id can be either busName of the SNI, or sender of the SNI at registration time
func (s *statusNotifierWatcher) getItem(id string) (*statusNotifierItem, bool) {
	s.RLock()
	defer s.RUnlock()
	if item, ok := s.items[id]; ok {
		return item, ok
	}
	item, ok := s.itemSenders[id]
	return item, ok
}

func (s *statusNotifierWatcher) init() error {
	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniName),
		dbus.WithMatchMember(sniMemberNewTitle),
	); err != nil {
		return err
	}
	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniName),
		dbus.WithMatchMember(sniMemberNewIcon),
	); err != nil {
		return err
	}
	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniName),
		dbus.WithMatchMember(sniMemberNewAttentionIcon),
	); err != nil {
		return err
	}
	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniName),
		dbus.WithMatchMember(sniMemberNewOverlayIcon),
	); err != nil {
		return err
	}
	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniName),
		dbus.WithMatchMember(sniMemberNewToolTip),
	); err != nil {
		return err
	}
	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniName),
		dbus.WithMatchMember(sniMemberNewToolTip),
	); err != nil {
		return err
	}

	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniMenuName),
		dbus.WithMatchMember(sniMenuMemberItemPropertiesUpdated),
	); err != nil {
		return err
	}
	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(sniMenuName),
		dbus.WithMatchMember(sniMenuMemberLayoutUpdated),
	); err != nil {
		return err
	}

	if err := s.conn.AddMatchSignal(
		dbus.WithMatchInterface(fdoName),
		dbus.WithMatchObjectPath(fdoPath),
	); err != nil {
		return err
	}

	reply, err := s.conn.RequestName(snwName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner && reply != dbus.RequestNameReplyAlreadyOwner {
		return fmt.Errorf("DBUS StatusNotifierWatcher already claimed, disable systray or close the other claiming application: code %d", reply)
	}

	if err := s.conn.Export(s, snwPath, snwName); err != nil {
		return err
	}

	propsSpec := map[string]map[string]*prop.Prop{
		snwName: {
			`RegisteredStatusNotifierItems`: {
				Value:    s.RegisteredStatusNotifierItems(),
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			`IsStatusNotifierHostRegistered`: {
				Value:    s.IsStatusNotifierHostRegistered(),
				Writable: false,
				Emit:     prop.EmitFalse,
			},
			`ProtocolVersion`: {
				Value:    s.ProtocolVersion(),
				Writable: false,
				Emit:     prop.EmitConst,
			},
		},
	}
	s.props, err = prop.Export(s.conn, snwPath, propsSpec)
	if err != nil {
		return err
	}

	snwIface, err := ifaces.ReadFile(`interfaces/org.kde.StatusNotifierWatcher.xml`)
	if err != nil {
		return err
	}
	if err := s.conn.Export(introspect.Introspectable(snwIface), snwPath, fdoIntrospectableName); err != nil {
		return err
	}

	sniIface, err := ifaces.ReadFile(`interfaces/org.kde.StatusNotifierItem.xml`)
	if err != nil {
		return err
	}
	if err := s.conn.Export(introspect.Introspectable(sniIface), sniPath, sniName); err != nil {
		return err
	}

	menuIface, err := ifaces.ReadFile(`interfaces/com.canonical.dbusmenu.xml`)
	if err != nil {
		return err
	}
	if err := s.conn.Export(introspect.Introspectable(menuIface), sniMenuPath, sniMenuName); err != nil {
		return err
	}

	if err := s.conn.Emit(snwPath, snwSignalHostRegistered); err != nil {
		s.log.Warn(`Failed emitting signal`, `path`, snwPath, `signal`, snwSignalHostRegistered)
		return err
	}

	go s.watch()

	return nil
}

func (s *statusNotifierWatcher) watch() {
	for {
		select {
		case <-s.quitCh:
			return
		default:
			select {
			case <-s.quitCh:
				return
			case sig, ok := <-s.signals:
				if !ok {
					return
				}

				s.log.Trace(`Received dbus signal`, `sig`, sig.Name, `sender`, sig.Sender, `objectPath`, sig.Path)

				switch sig.Name {
				case sniSignalNewTitle:
					item, ok := s.getItem(sig.Sender)
					if !ok {
						s.log.Debug(`Received signal for unknown item`, `sender`, sig.Sender, `sig`, sig.Name)
						continue
					}
					if err := item.updateTitle(); err != nil {
						s.log.Debug(`Failed updating item`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}

					data := &eventv1.UpdateTitleValue{
						BusName: item.busName,
						Title:   item.target.Title,
					}
					anyData, err := anypb.New(data)
					if err != nil {
						s.log.Debug(`Failed encoding event`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}
					s.eventCh <- &eventv1.Event{
						Kind: eventv1.EventKind_EVENT_KIND_DBUS_UPDATETITLE,
						Data: anyData,
					}
				case sniSignalNewToolTip:
					item, ok := s.getItem(sig.Sender)
					if !ok {
						s.log.Debug(`Received signal for unknown item`, `sender`, sig.Sender, `sig`, sig.Name)
						continue
					}
					if err := item.updateTooltip(); err != nil {
						s.log.Debug(`Failed updating item`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}

					data := &eventv1.UpdateTooltipValue{
						BusName: item.busName,
						Tooltip: item.target.Tooltip,
					}
					anyData, err := anypb.New(data)
					if err != nil {
						s.log.Debug(`Failed encoding event`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}
					s.eventCh <- &eventv1.Event{
						Kind: eventv1.EventKind_EVENT_KIND_DBUS_UPDATETOOLTIP,
						Data: anyData,
					}
				case sniSignalNewIcon, sniSignalNewAttentionIcon, sniSignalNewOverlayIcon, sniSignalNewStatus:
					item, ok := s.getItem(sig.Sender)
					if !ok {
						s.log.Debug(`Received signal for unknown item`, `sender`, sig.Sender, `sig`, sig.Name)
						continue
					}

					if err := item.updateStatus(); err != nil {
						s.log.Debug(`Failed updating item`, `busName`, item.busName, `sig`, sig.Name, `err`, err)

					}
					if err := item.updateIcon(); err != nil {
						s.log.Debug(`Failed updating item`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}

					data := &eventv1.UpdateIconValue{
						BusName: item.busName,
						Icon:    item.target.Icon,
					}
					anyData, err := anypb.New(data)
					if err != nil {
						s.log.Debug(`Failed encoding event`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}
					s.eventCh <- &eventv1.Event{
						Kind: eventv1.EventKind_EVENT_KIND_DBUS_UPDATEICON,
						Data: anyData,
					}

					if sig.Name == sniSignalNewStatus {
						data := &eventv1.UpdateStatusValue{
							BusName: item.busName,
							Status:  item.target.Status,
						}
						anyData, err := anypb.New(data)
						if err != nil {
							s.log.Debug(`Failed encoding event`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
							continue
						}
						s.eventCh <- &eventv1.Event{
							Kind: eventv1.EventKind_EVENT_KIND_DBUS_UPDATESTATUS,
							Data: anyData,
						}
					}
				case sniMenuSignalLayoutUpdated, sniMenuSignalItemsPropertiesUpdated:
					item, ok := s.getItem(sig.Sender)
					if !ok {
						s.log.Debug(`Received signal for unknown item`, `sender`, sig.Sender, `sig`, sig.Name)
						continue
					}
					if err := item.updateMenu(); err != nil {
						s.log.Debug(`Failed updating item`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}

					data := &eventv1.UpdateMenuValue{
						BusName: item.busName,
						Menu:    item.target.Menu,
					}
					anyData, err := anypb.New(data)
					if err != nil {
						s.log.Debug(`Failed encoding event`, `busName`, item.busName, `sig`, sig.Name, `err`, err)
						continue
					}
					s.eventCh <- &eventv1.Event{
						Kind: eventv1.EventKind_EVENT_KIND_DBUS_UPDATEMENU,
						Data: anyData,
					}
				case fdoSignalNameOwnerChanged:
					if len(sig.Body) != 3 {
						s.log.Debug(`Malformed event`, `busName`, sig.Sender, `sig`, sig.Name)
						continue
					}
					name, ok := sig.Body[0].(string)
					if !ok {
						s.log.Debug(`Malformed event`, `busName`, sig.Sender, `sig`, sig.Name)
						continue
					}
					newOwner, ok := sig.Body[2].(string)
					if !ok {
						s.log.Debug(`Malformed event`, `busName`, sig.Sender, `sig`, sig.Name)
						continue
					}

					item, ok := s.getItem(name)
					if !ok {
						s.log.Trace(`Received signal for unknown item`, `busName`, name, `sig`, sig.Name)
						continue
					}

					if newOwner != `` {
						if newOwner != name {
							s.Lock()
							s.items[newOwner] = item
							delete(s.items, name)
							s.Unlock()
						}
						continue
					}

					evt, err := eventv1.NewString(eventv1.EventKind_EVENT_KIND_DBUS_UNREGISTERSTATUSNOTIFIER, item.busName)
					if err != nil {
						s.log.Debug(`Failed encoding event`, `busName`, name, `sig`, sig.Name, `err`, err)
						continue
					}
					s.eventCh <- evt
					s.Lock()
					delete(s.items, item.busName)
					delete(s.itemSenders, name)
					s.Unlock()
					if err := s.conn.Emit(snwPath, snwSignalItemUnregistered, fmt.Sprintf("%s%s", item.busName, item.objectPath)); err != nil {
						s.log.Warn(`Failed emitting unregister signal for sni`, `busName`, item.busName, `objectPath`, item.objectPath, `err`, err)
					}
				}
			}
		}
	}
}

func (s *statusNotifierWatcher) close() {
	close(s.quitCh)
}

func newStatusNotifierWatcher(conn *dbus.Conn, logger hclog.Logger, eventCh chan *eventv1.Event) (*statusNotifierWatcher, error) {
	s := &statusNotifierWatcher{
		conn:        conn,
		log:         logger,
		items:       make(map[string]*statusNotifierItem),
		itemSenders: make(map[string]*statusNotifierItem),
		eventCh:     eventCh,
		signals:     make(chan *dbus.Signal, 10),
		quitCh:      make(chan struct{}),
	}

	s.conn.Signal(s.signals)

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
