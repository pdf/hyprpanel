package dbus

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/hashicorp/go-hclog"
)

const (
	portalName = `org.freedesktop.portal.Desktop`
	portalPath = dbus.ObjectPath(`/org/freedesktop/portal/desktop`)

	portalRequestName           = `org.freedesktop.portal.Request`
	portalRequestMemberResponse = `Response`
	portalRequestSignalResponse = portalRequestName + `.` + portalRequestMemberResponse

	portalSessionName        = `org.freedesktop.portal.Session`
	portalSessionMethodClose = portalSessionName + `.Close`

	portalTokenPrefix = `hyprpanel_`
)

type portalClient struct {
	sync.RWMutex
	conn     *dbus.Conn
	log      hclog.Logger
	requests map[dbus.ObjectPath]chan *dbus.Signal
	busObj   dbus.BusObject
	signals  chan *dbus.Signal
	quitCh   chan struct{}
}

func (c *portalClient) request(ctx context.Context, method string, token string, args ...any) (*dbus.Signal, error) {
	requestPath := c.requestPath(token)
	responseCh := make(chan *dbus.Signal, 1)

	c.Lock()
	c.requests[requestPath] = responseCh
	c.Unlock()

	defer func() {
		c.Lock()
		delete(c.requests, requestPath)
		c.Unlock()
	}()

	if call := c.busObj.Call(method, 0, args...); call.Err != nil {
		return nil, call.Err
	}

	select {
	case sig := <-responseCh:
		return sig, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *portalClient) requestPath(token string) dbus.ObjectPath {
	busName := c.conn.Names()[0]
	busName = strings.ReplaceAll(busName, `:`, ``)
	busName = strings.ReplaceAll(busName, `.`, `_`)
	return dbus.ObjectPath(fmt.Sprintf("/org/freedesktop/portal/desktop/request/%s/%s", busName, token))
}

func (c *portalClient) token() string {
	return portalTokenPrefix + strconv.Itoa(rand.Int())
}

func (c *portalClient) watch() {
	for {
		select {
		case <-c.quitCh:
			return
		default:
			select {
			case <-c.quitCh:
				return
			case sig, ok := <-c.signals:
				if !ok {
					return
				}
				switch sig.Name {
				case portalRequestSignalResponse:
					c.RLock()
					if ch, ok := c.requests[sig.Path]; ok {
						ch <- sig
					}
					c.RUnlock()
				}
			}
		}
	}
}

func newPortalClient(conn *dbus.Conn, logger hclog.Logger) (*portalClient, error) {
	c := &portalClient{
		conn:     conn,
		log:      logger,
		busObj:   conn.Object(portalName, portalPath),
		requests: make(map[dbus.ObjectPath]chan *dbus.Signal),
		signals:  make(chan *dbus.Signal),
		quitCh:   make(chan struct{}),
	}

	if err := c.conn.AddMatchSignal(
		dbus.WithMatchInterface(portalRequestName),
		dbus.WithMatchMember(portalRequestMemberResponse),
	); err != nil {
		return nil, err
	}

	c.conn.Signal(c.signals)

	go c.watch()

	return c, nil
}
