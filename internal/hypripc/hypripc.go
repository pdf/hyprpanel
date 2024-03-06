// Package hypripc provides an API for interacting with the Hyprland IPC bus
package hypripc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

// Event enum.
type Event string

const (
	// EventUnspecified is the default catch-all event.
	EventUnspecified = `unspecified`
	// EventWorkspace event identifier.
	EventWorkspace = `workspace`
	// EventFocusedMon event identifier.
	EventFocusedMon = `focusedmon`
	// EventActiveWindow event identifier.
	EventActiveWindow = `activewindow`
	// EventActiveWindowV2 event identifier.
	EventActiveWindowV2 = `activewindowv2`
	// EventFullscreen event identifier.
	EventFullscreen = `fullscreen`
	// EventMonitorRemoved event identifier.
	EventMonitorRemoved = `monitorremoved`
	// EventMonitorAdded event identifier.
	EventMonitorAdded = `monitoradded`
	// EventCreateWorkspace event identifier.
	EventCreateWorkspace = `createworkspace`
	// EventDestroyWorkspace event identifier.
	EventDestroyWorkspace = `destroyworkspace`
	// EventMoveWorkspace event identifier.
	EventMoveWorkspace = `moveworkspace`
	// EventRenameWorkspace event identifier.
	EventRenameWorkspace = `renameworkspace`
	// EventActiveSpecial event identifier.
	EventActiveSpecial = `activespecial`
	// EventActiveLayout event identifier.
	EventActiveLayout = `activelayout`
	// EventOpenWindow event identifier.
	EventOpenWindow = `openwindow`
	// EventCloseWindow event identifier.
	EventCloseWindow = `closewindow`
	// EventMoveWindow event identifier.
	EventMoveWindow = `movewindow`
	// EventOpenLayer event identifier.
	EventOpenLayer = `openlayer`
	// EventCloseLayer event identifier.
	EventCloseLayer = `closelayer`
	// EventSubmap event identifier.
	EventSubmap = `submap`
	// EventChangeFloatingMode event identifier.
	EventChangeFloatingMode = `changefloatingmode`
	// EventUrgent event identifier.
	EventUrgent = `urgent`
	// EventMinimize event identifier.
	EventMinimize = `minimize`
	// EventScreencast event identifier.
	EventScreencast = `screencast`
	// EventWindowTitle event identifier.
	EventWindowTitle = `windowtitle`
	// EventIgnoreGroupLock event identifier.
	EventIgnoreGroupLock = `ignoregrouplock`
	// EventLockGroups event identifier.
	EventLockGroups = `lockgroups`

	// DispatchWorkspace dispatcher identifier.
	DispatchWorkspace = `workspace`
	// DispatchFocusWindow dispatcher identifier.
	DispatchFocusWindow = `focuswindow`
	// DispatchCloseWindow dispatcher identifier.
	DispatchCloseWindow = `closewindow`
)

var (
	eventMatch = regexp.MustCompile(`^(?P<Event>[^>]+)>>(?P<Value>.*)$`)
)

// CancelFunc cancels a subscription when called.
type CancelFunc func()

// HyprIPC client.
type HyprIPC struct {
	log           hclog.Logger
	subscriptions map[Event]map[uuid.UUID]chan *eventv1.Event
	evtConn       net.Conn
	evtBus        chan []byte
	mu            sync.RWMutex
}

// ActiveWindow returns the currently active window client.
func (h *HyprIPC) ActiveWindow() (*Client, error) {
	res, err := h.send(`activewindow`)
	if err != nil {
		return nil, err
	}

	client := &Client{}
	if err := json.Unmarshal(res, client); err != nil {
		return nil, err
	}

	return client, nil
}

// ActiveWorkspace returns the currently actie workspace.
func (h *HyprIPC) ActiveWorkspace() (*Workspace, error) {
	res, err := h.send(`activeworkspace`)
	if err != nil {
		return nil, err
	}

	workspace := &Workspace{}
	if err := json.Unmarshal(res, workspace); err != nil {
		return nil, err
	}

	return workspace, nil
}

// Clients returns a list of all active client windows,
func (h *HyprIPC) Clients() ([]Client, error) {
	res, err := h.send(`clients`)
	if err != nil {
		return nil, err
	}

	clients := make([]Client, 0)
	if err := json.Unmarshal(res, &clients); err != nil {
		return nil, err
	}

	return clients, nil
}

// Monitors returns a list of active monitors.
func (h *HyprIPC) Monitors() ([]Monitor, error) {
	res, err := h.send(`monitors all`)
	if err != nil {
		return nil, err
	}

	monitors := make([]Monitor, 0)
	if err := json.Unmarshal(res, &monitors); err != nil {
		return nil, err
	}

	return monitors, nil
}

// Workspaces returns a list of all active workspaces.
func (h *HyprIPC) Workspaces() ([]Workspace, error) {
	res, err := h.send(`workspaces`)
	if err != nil {
		return nil, err
	}

	workspaces := make([]Workspace, 0)
	if err := json.Unmarshal(res, &workspaces); err != nil {
		return nil, err
	}

	return workspaces, nil
}

// Dispatch calls a dispatcher.
func (h *HyprIPC) Dispatch(args ...string) error {
	_, err := h.send(append([]string{`dispatch`}, args...)...)
	return err
}

func (h *HyprIPC) send(args ...string) ([]byte, error) {
	ctrl, err := net.Dial(`unix`, fmt.Sprintf("/tmp/hypr/%s/.socket.sock", os.Getenv(`HYPRLAND_INSTANCE_SIGNATURE`)))
	if err != nil {
		return nil, err
	}
	defer ctrl.Close()

	if _, err := ctrl.Write([]byte(fmt.Sprintf("j/%s", strings.Join(args, ` `)))); err != nil {
		return nil, err
	}

	return io.ReadAll(ctrl)
}

// Subscribe returns a channel that will emit the specified event(s) when they arrive.
func (h *HyprIPC) Subscribe(evt ...Event) (chan *eventv1.Event, CancelFunc) {
	id := uuid.New()
	ch := make(chan *eventv1.Event)

	h.mu.Lock()
	defer h.mu.Unlock()

	if len(evt) == 0 {
		evt = []Event{EventUnspecified}
	}

	for _, e := range evt {
		if _, ok := h.subscriptions[e]; !ok {
			h.subscriptions[e] = make(map[uuid.UUID]chan *eventv1.Event, 0)
		}

		h.subscriptions[e][id] = ch
	}

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		for _, e := range evt {
			delete(h.subscriptions[e], id)
		}
		close(ch)
	}
}

// StartEvents begins the event loop.
func (h *HyprIPC) StartEvents() {
	go h.readloop()
	go h.eventloop()
}

// Close terminates all connections, event loops, and closes all subscriptions.
func (h *HyprIPC) Close() {
	h.evtConn.Close()
	close(h.evtBus)
}

func (h *HyprIPC) eventloop() {
	for line := range h.evtBus {
		result := eventMatch.FindSubmatch(line)
		if len(result) < 3 {
			continue
		}

		name := Event(result[1])
		value := string(result[2])
		h.log.Trace(`Received hypr msg`, `name`, string(name), `value`, value)
		evt, err := hyprToEvent(name, value)
		if err != nil {
			h.log.Warn(`failed parsing hyprland event`, `err`, err)
			continue
		}
		h.mu.RLock()
		if _, ok := h.subscriptions[EventUnspecified]; ok {
			for _, ch := range h.subscriptions[EventUnspecified] {
				ch <- evt
			}
		}
		if _, ok := h.subscriptions[name]; !ok {
			h.mu.RUnlock()
			continue
		}
		for _, ch := range h.subscriptions[name] {
			ch <- evt
		}
		h.mu.RUnlock()
	}
}

func (h *HyprIPC) readloop() {
	scanner := bufio.NewScanner(h.evtConn)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		h.evtBus <- scanner.Bytes()
	}

	if scanner.Err() != nil {
		h.log.Error(`Lost connection to hyprland IPC bus`, `err`, scanner.Err())
	}
}

// New instantiates a new HyprIPC client
func New(log hclog.Logger) (*HyprIPC, error) {
	evtConn, err := net.Dial(`unix`, fmt.Sprintf("/tmp/hypr/%s/.socket2.sock", os.Getenv(`HYPRLAND_INSTANCE_SIGNATURE`)))
	if err != nil {
		return nil, err
	}

	ipc := &HyprIPC{
		log:           log,
		evtConn:       evtConn,
		evtBus:        make(chan []byte, 10),
		subscriptions: make(map[Event]map[uuid.UUID]chan *eventv1.Event),
	}

	return ipc, nil
}

func hyprToEvent(name Event, value string) (*eventv1.Event, error) {
	switch name {
	case EventWorkspace:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_WORKSPACE, value)
	case EventFocusedMon:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_FOCUSEDMON, value)
	case EventActiveWindow:
		s := strings.SplitN(value, `,`, 4)
		if len(s) != 2 {
			return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOW, value)
		}
		data, err := anypb.New(&eventv1.HyprActiveWindowValue{Class: s[0], Title: s[1]})
		if err != nil {
			return nil, fmt.Errorf("invalid event (%s): %w", EventActiveWindow, err)
		}
		return &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOW,
			Data: data,
		}, nil
	case EventActiveWindowV2:
		value = `0x` + value
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOWV2, value)
	case EventFullscreen:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_FULLSCREEN, value)
	case EventMonitorRemoved:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MONITORREMOVED, value)
	case EventMonitorAdded:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MONITORADDED, value)
	case EventCreateWorkspace:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CREATEWORKSPACE, value)
	case EventDestroyWorkspace:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_DESTROYWORKSPACE, value)
	case EventMoveWorkspace:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MOVEWORKSPACE, value)
	case EventRenameWorkspace:
		s := strings.SplitN(value, `,`, 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("invalid event (%s)", EventRenameWorkspace)
		}
		id, err := strconv.Atoi(s[0])
		if err != nil {
			return nil, fmt.Errorf("invalid event (%s): %w", EventRenameWorkspace, err)
		}
		data, err := anypb.New(&eventv1.HyprRenameWorkspaceValue{Id: int32(id), Name: s[1]})
		if err != nil {
			return nil, fmt.Errorf("invalid event (%s): %w", EventRenameWorkspace, err)
		}
		return &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_HYPR_RENAMEWORKSPACE,
			Data: data,
		}, nil
	case EventActiveSpecial:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVESPECIAL, value)
	case EventActiveLayout:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVELAYOUT, value)
	case EventOpenWindow:
		s := strings.SplitN(value, `,`, 4)
		if len(s) != 4 {
			return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_OPENWINDOW, value)
		}
		data, err := anypb.New(&eventv1.HyprOpenWindowValue{
			Address:       s[0],
			WorkspaceName: s[1],
			Class:         s[2],
			Title:         s[3],
		})
		if err != nil {
			return nil, fmt.Errorf("invalid event (%s): %w", EventOpenWindow, err)
		}
		return &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_HYPR_OPENWINDOW,
			Data: data,
		}, nil
	case EventCloseWindow:
		value = `0x` + value
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CLOSEWINDOW, value)
	case EventMoveWindow:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MOVEWINDOW, value)
	case EventOpenLayer:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_OPENLAYER, value)
	case EventCloseLayer:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CLOSELAYER, value)
	case EventSubmap:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_SUBMAP, value)
	case EventChangeFloatingMode:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CHANGEFLOATINGMODE, value)
	case EventUrgent:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_URGENT, value)
	case EventMinimize:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MINIMIZE, value)
	case EventScreencast:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_SCREENCAST, value)
	case EventWindowTitle:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_WINDOWTITLE, value)
	case EventIgnoreGroupLock:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_IGNOREGROUPLOCK, value)
	case EventLockGroups:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_EVENTLOCKGROUPS, value)
	default:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_UNSPECIFIED, value)
	}
}
