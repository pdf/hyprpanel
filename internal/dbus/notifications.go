package dbus

import (
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/hashicorp/go-hclog"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// NotificationHintKey enum
type NotificationHintKey string

const (
	notificationsName = `org.freedesktop.Notifications`
	notificationsPath = `/org/freedesktop/Notifications`

	notificationsSignalNotificationClosed = notificationsName + `.NotificationClosed`
	notificationsSignalActionInvoked      = notificationsName + `.ActionInvoked`

	notificationInfoName        = `com.c0dedbad.hyprpanel`
	notificationInfoVendor      = `pdf`
	notificationInfoSpecVersion = `1.2`

	// NotificationHintKeyActionIcons BOOLEAN	When set, a server that has the "action-icons" capability will attempt to interpret any action identifier as a named icon. The localized display name will be used to annotate the icon for accessibility purposes. The icon name should be compliant with the Freedesktop.org Icon Naming Specification.	>= 1.2
	NotificationHintKeyActionIcons NotificationHintKey = "action-icons"
	// NotificationHintKeyCategory STRING	The type of notification this is.
	NotificationHintKeyCategory NotificationHintKey = "category"
	// NotificationHintKeyDesktopEntry STRING	This specifies the name of the desktop filename representing the calling program. This should be the same as the prefix used for the application's .desktop file. An example would be "rhythmbox" from "rhythmbox.desktop". This can be used by the daemon to retrieve the correct icon for the application, for logging purposes, etc.
	NotificationHintKeyDesktopEntry NotificationHintKey = "desktop-entry"
	// NotificationHintKeyImageData (iiibiiay)	This is a raw data image format which describes the width, height, rowstride, has alpha, bits per sample, channels and image data respectively.	>= 1.2
	NotificationHintKeyImageData NotificationHintKey = "image-data"
	// NotificationHintKeyImageDataAlt (iiibiiay)	Deprecated. Use image-data instead.	= 1.1
	NotificationHintKeyImageDataAlt NotificationHintKey = "image_data"
	// NotificationHintKeyImagePath STRING	Alternative way to define the notification image. See Icons and Images.	>= 1.2
	NotificationHintKeyImagePath NotificationHintKey = "image-path"
	// NotificationHintKeyImagePathAlt STRING	Deprecated. Use image-path instead.	= 1.1
	NotificationHintKeyImagePathAlt NotificationHintKey = "image_path"
	// NotificationHintKeyIconDataAlt (iiibiiay)	Deprecated. Use image-data instead.	< 1.1
	NotificationHintKeyIconDataAlt NotificationHintKey = "icon_data"
	// NotificationHintKeyResident BOOLEAN	When set the server will not automatically remove the notification when an action has been invoked. The notification will remain resident in the server until it is explicitly removed by the user or by the sender. This hint is likely only useful when the server has the "persistence" capability.	>= 1.2
	NotificationHintKeyResident NotificationHintKey = "resident"
	// NotificationHintKeySoundFile STRING	The path to a sound file to play when the notification pops up.
	NotificationHintKeySoundFile NotificationHintKey = "sound-file"
	// NotificationHintKeySoundName STRING	A themeable named sound from the freedesktop.org sound naming specification to play when the notification pops up. Similar to icon-name, only for sounds. An example would be "message-new-instant".
	NotificationHintKeySoundName NotificationHintKey = "sound-name"
	// NotificationHintKeySuppressSound BOOLEAN	Causes the server to suppress playing any sounds, if it has that ability. This is usually set when the client itself is going to play its own sound.
	NotificationHintKeySuppressSound NotificationHintKey = "suppress-sound"
	// NotificationHintKeyTransient BOOLEAN	When set the server will treat the notification as transient and by-pass the server's persistence capability, if it should exist.	>= 1.2
	NotificationHintKeyTransient NotificationHintKey = "transient"
	// NotificationHintKeyX INT32	Specifies the X location on the screen that the notification should point to. The "y" hint must also be specified.
	NotificationHintKeyX NotificationHintKey = "x"
	// NotificationHintKeyY INT32	Specifies the Y location on the screen that the notification should point to. The "x" hint must also be specified.
	NotificationHintKeyY NotificationHintKey = "y"
	// NotificationHintKeyUrgency BYTE|UINT32	The urgency level.
	NotificationHintKeyUrgency NotificationHintKey = "urgency"
	// NotificationHintKeySenderPid NON-STANDARD INT64	process id of the sender
	NotificationHintKeySenderPid NotificationHintKey = "sender-pid"
)

/*
Capabilities

"action-icons"	Supports using icons instead of text for displaying actions. Using icons for actions must be enabled on a per-notification basis using the "action-icons" hint.
"actions"	The server will provide the specified actions to the user. Even if this cap is missing, actions may still be specified by the client, however the server is free to ignore them.
"body"	Supports body text. Some implementations may only show the summary (for instance, onscreen displays, marquee/scrollers)
"body-hyperlinks"	The server supports hyperlinks in the notifications.
"body-images"	The server supports images in the notifications.
"body-markup"	Supports markup in the body text. If marked up text is sent to a server that does not give this cap, the markup will show through as regular text so must be stripped clientside.
"icon-multi"	The server will render an animation of all the frames in a given image array. The client may still specify multiple frames even if this cap and/or "icon-static" is missing, however the server is free to ignore them and use only the primary frame.
"icon-static"	Supports display of exactly 1 frame of any given image array. This value is mutually exclusive with "icon-multi", it is a protocol error for the server to specify both.
"persistence"	The server supports persistence of notifications. Notifications will be retained until they are acknowledged or removed by the user or recalled by the sender. The presence of this capability allows clients to depend on the server to ensure a notification is seen and eliminate the need for the client to display a reminding function (such as a status icon) of its own.
"sound"	The server supports sounds on notifications. If returned, the server must support the "sound-file" and "suppress-sound" hints.
*/
var notificationCapabilities = []string{`action-icons`, `actions`, `body`, `body-hyperlinks`, `body-images`, `body-markup`, `icon-static`, `persistence`}

type notifications struct {
	sync.RWMutex
	conn   *dbus.Conn
	log    hclog.Logger
	lastID atomic.Uint32

	eventCh chan *eventv1.Event
}

func (n *notifications) Notify(appName string, replacesID uint32, appIcon string, summary string, body string, actions []string, hints map[NotificationHintKey]dbus.Variant, timeout int32) (uint32, *dbus.Error) {
	id := n.lastID.Add(1)
	if len(actions)%2 != 0 {
		return 0, &dbus.ErrMsgInvalidArg
	}

	n.log.Trace(`Received notification`, `appName`, appName, `replacesID`, replacesID, `appIcon`, appIcon, `summary`, summary, `body`, body, `actions`, actions, `hints`, hints, `timeout`, timeout)

	notification := &eventv1.NotificationValue{
		Id:         id,
		AppName:    appName,
		ReplacesId: replacesID,
		AppIcon:    appIcon,
		Summary:    summary,
		Body:       body,
		Actions:    make([]*eventv1.NotificationValue_Action, len(actions)/2),
		Hints:      make([]*eventv1.NotificationValue_Hint, len(hints)),
		Timeout:    durationpb.New(time.Duration(timeout) * time.Millisecond),
	}

	var actionKey string
	for i, v := range actions {
		if i%2 == 0 {
			actionKey = v
			continue
		}
		notification.Actions[i/2] = &eventv1.NotificationValue_Action{
			Key:   actionKey,
			Value: v,
		}
	}

	i := 0
	for k, v := range hints {
		val, err := hintToAny(k, v)
		if err != nil {
			n.log.Debug(`Skipping unknown notification hint`, `name`, k, `value`, v, `err`, err)
		}
		notification.Hints[i] = &eventv1.NotificationValue_Hint{
			Key:   string(k),
			Value: val,
		}
		i++
	}

	data, err := anypb.New(notification)
	if err != nil {
		return 0, &dbus.ErrMsgInvalidArg
	}
	n.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_DBUS_NOTIFICATION,
		Data: data,
	}

	return id, nil
}

func (n *notifications) CloseNotification(id uint32) *dbus.Error {
	idpb := wrapperspb.UInt32(id)
	data, err := anypb.New(idpb)
	if err != nil {
		return &dbus.ErrMsgInvalidArg
	}
	n.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_DBUS_CLOSENOTIFICATION,
		Data: data,
	}
	n.log.Trace(`Emitting notification closed signal`, `id`, id, `reason`, 0)
	if err := n.conn.Emit(notificationsPath, notificationsSignalNotificationClosed, id, 0); err != nil {
		return &dbus.ErrMsgInvalidArg
	}
	return nil
}

func (n *notifications) GetCapabilities() ([]string, *dbus.Error) {
	return notificationCapabilities, nil
}

func (n *notifications) GetServerInformation() (name, vendor, version, specVersion string, derr *dbus.Error) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ``, ``, ``, ``, &dbus.ErrMsgInvalidArg
	}
	return notificationInfoName, notificationInfoVendor, buildInfo.Main.Version, notificationInfoSpecVersion, nil
}

func (n *notifications) init() error {
	reply, err := n.conn.RequestName(notificationsName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner && reply != dbus.RequestNameReplyAlreadyOwner {
		return fmt.Errorf("DBUS Notifications already claimed, disable systray or close the other claiming application: code %d", reply)
	}

	if err := n.conn.Export(n, notificationsPath, notificationsName); err != nil {
		return err
	}

	notificationsIface, err := ifaces.ReadFile(`interfaces/org.freedesktop.Notifications.xml`)
	if err != nil {
		return err
	}
	if err := n.conn.Export(introspect.Introspectable(notificationsIface), notificationsPath, fdoIntrospectableName); err != nil {
		return err
	}

	return nil

}

func (n *notifications) Closed(id uint32, reason hyprpanelv1.NotificationClosedReason) error {
	n.log.Trace(`Emitting notification closed signal`, `id`, id, `reason`, reason)
	if err := n.conn.Emit(notificationsPath, notificationsSignalNotificationClosed, id, reason); err != nil {
		return &dbus.ErrMsgInvalidArg
	}
	return nil
}

func (n *notifications) Action(id uint32, actionKey string) error {
	n.log.Trace(`Emitting notification action invoked signal`, `id`, id, `actionKey`, actionKey)
	if err := n.conn.Emit(notificationsPath, notificationsSignalActionInvoked, id, actionKey); err != nil {
		return &dbus.ErrMsgInvalidArg
	}
	return nil
}

func newNotifications(conn *dbus.Conn, logger hclog.Logger, eventCh chan *eventv1.Event) (*notifications, error) {
	n := &notifications{
		conn:    conn,
		log:     logger,
		eventCh: eventCh,
	}

	if err := n.init(); err != nil {
		return nil, err
	}

	return n, nil
}

func hintToAny(name NotificationHintKey, val dbus.Variant) (*anypb.Any, error) {
	if val.Signature().Empty() {
		return nil, nil
	}
	switch name {
	case NotificationHintKeyCategory, NotificationHintKeyDesktopEntry, NotificationHintKeyImagePath, NotificationHintKeyImagePathAlt, NotificationHintKeySoundFile, NotificationHintKeySoundName:
		var v string
		if err := val.Store(&v); err != nil {
			return nil, err
		}
		return anypb.New(wrapperspb.String(v))
	case NotificationHintKeyActionIcons, NotificationHintKeyResident, NotificationHintKeySuppressSound, NotificationHintKeyTransient:
		var v bool
		if err := val.Store(&v); err != nil {
			return nil, err
		}
		return anypb.New(wrapperspb.Bool(v))
	case NotificationHintKeyX, NotificationHintKeyY:
		var v int32
		if err := val.Store(&v); err != nil {
			return nil, err
		}
		return anypb.New(wrapperspb.Int32(v))
	case NotificationHintKeyImageData, NotificationHintKeyImageDataAlt, NotificationHintKeyIconDataAlt:
		var v = &eventv1.NotificationValue_Pixmap{}
		if err := val.Store(v); err != nil {
			return nil, err
		}
		return anypb.New(v)
	case NotificationHintKeyUrgency:
		// Urgency is non-standard, some callers us byte, some uint32, normalize to uint32.
		var v byte
		if err := val.Store(&v); err != nil {
			var u uint32
			if err := val.Store(&v); err != nil {
				return nil, err
			}
			return anypb.New(wrapperspb.UInt32(u))
		}
		return anypb.New(wrapperspb.UInt32(uint32(v)))
	case NotificationHintKeySenderPid:
		var v int64
		if err := val.Store(&v); err != nil {
			return nil, err
		}
		return anypb.New(wrapperspb.Int64(v))
	default:
		return nil, fmt.Errorf(`unhandled hint: %s (%s)`, name, val.Signature().String())
	}
}
