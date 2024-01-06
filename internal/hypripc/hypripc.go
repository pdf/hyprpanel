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

type Event string

const (
	EventUnspecified        = `unspecified`
	EventWorkspace          = `workspace`
	EventFocusedmon         = `focusedmon`
	EventActivewindow       = `activewindow`
	EventActivewindowv2     = `activewindowv2`
	EventFullscreen         = `fullscreen`
	EventMonitorremoved     = `monitorremoved`
	EventMonitoradded       = `monitoradded`
	EventCreateworkspace    = `createworkspace`
	EventDestroyworkspace   = `destroyworkspace`
	EventMoveworkspace      = `moveworkspace`
	EventRenameworkspace    = `renameworkspace`
	EventActivespecial      = `activespecial`
	EventActivelayout       = `activelayout`
	EventOpenwindow         = `openwindow`
	EventClosewindow        = `closewindow`
	EventMovewindow         = `movewindow`
	EventOpenlayer          = `openlayer`
	EventCloselayer         = `closelayer`
	EventSubmap             = `submap`
	EventChangefloatingmode = `changefloatingmode`
	EventUrgent             = `urgent`
	EventMinimize           = `minimize`
	EventScreencast         = `screencast`
	EventWindowtitle        = `windowtitle`
	EventIgnoregrouplock    = `ignoregrouplock`
	EventLockgroups         = `lockgroups`

	DispatchWorkspace   = `workspace`
	DispatchFocusWindow = `focuswindow`
	DispatchCloseWindow = `closewindow`
)

var (
	eventMatch = regexp.MustCompile(`^(?P<Event>[^>]+)>>(?P<Value>.*)$`)
)

type cancelFunc func()

type HyprIPC struct {
	log           hclog.Logger
	subscriptions map[Event]map[uuid.UUID]chan *eventv1.Event
	evtConn       net.Conn
	evtBus        chan []byte
	mu            sync.RWMutex
}

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

func (h *HyprIPC) Subscribe(evt Event) (chan *eventv1.Event, cancelFunc) {
	id := uuid.New()
	ch := make(chan *eventv1.Event)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscriptions[evt]; !ok {
		h.subscriptions[evt] = make(map[uuid.UUID]chan *eventv1.Event, 0)
	}

	h.subscriptions[evt][id] = ch

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.subscriptions[evt], id)
		close(ch)
	}
}

func (h *HyprIPC) StartEvents() {
	go h.readloop()
	go h.eventloop()
}

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
	case EventFocusedmon:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_FOCUSEDMON, value)
	case EventActivewindow:
		s := strings.SplitN(value, `,`, 4)
		if len(s) != 2 {
			return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOW, value)
		}
		data, err := anypb.New(&eventv1.HyprActiveWindowValue{Class: s[0], Title: s[1]})
		if err != nil {
			return nil, fmt.Errorf("invalid event (%s): %w", EventActivewindow, err)
		}
		return &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOW,
			Data: data,
		}, nil
	case EventActivewindowv2:
		value = `0x` + value
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVEWINDOWV2, value)
	case EventFullscreen:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_FULLSCREEN, value)
	case EventMonitorremoved:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MONITORREMOVED, value)
	case EventMonitoradded:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MONITORADDED, value)
	case EventCreateworkspace:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CREATEWORKSPACE, value)
	case EventDestroyworkspace:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_DESTROYWORKSPACE, value)
	case EventMoveworkspace:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MOVEWORKSPACE, value)
	case EventRenameworkspace:
		s := strings.SplitN(value, `,`, 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("invalid event (%s)", EventRenameworkspace)
		}
		id, err := strconv.Atoi(s[0])
		if err != nil {
			return nil, fmt.Errorf("invalid event (%s): %w", EventRenameworkspace, err)
		}
		data, err := anypb.New(&eventv1.HyprRenameWorkspaceValue{Id: int32(id), Name: s[1]})
		if err != nil {
			return nil, fmt.Errorf("invalid event (%s): %w", EventRenameworkspace, err)
		}
		return &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_HYPR_RENAMEWORKSPACE,
			Data: data,
		}, nil
	case EventActivespecial:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVESPECIAL, value)
	case EventActivelayout:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_ACTIVELAYOUT, value)
	case EventOpenwindow:
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
			return nil, fmt.Errorf("invalid event (%s): %w", EventOpenwindow, err)
		}
		return &eventv1.Event{
			Kind: eventv1.EventKind_EVENT_KIND_HYPR_OPENWINDOW,
			Data: data,
		}, nil
	case EventClosewindow:
		value = `0x` + value
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CLOSEWINDOW, value)
	case EventMovewindow:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MOVEWINDOW, value)
	case EventOpenlayer:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_OPENLAYER, value)
	case EventCloselayer:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CLOSELAYER, value)
	case EventSubmap:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_SUBMAP, value)
	case EventChangefloatingmode:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_CHANGEFLOATINGMODE, value)
	case EventUrgent:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_URGENT, value)
	case EventMinimize:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_MINIMIZE, value)
	case EventScreencast:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_SCREENCAST, value)
	case EventWindowtitle:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_WINDOWTITLE, value)
	case EventIgnoregrouplock:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_IGNOREGROUPLOCK, value)
	case EventLockgroups:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_HYPR_EVENTLOCKGROUPS, value)
	default:
		return eventv1.NewString(eventv1.EventKind_EVENT_KIND_UNSPECIFIED, value)
	}
}
