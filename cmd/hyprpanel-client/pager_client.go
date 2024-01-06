package main

import (
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pdf/hyprpanel/internal/hypripc"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/pdf/hyprpanel/style"
)

type pagerClient struct {
	*refTracker
	ws      *pagerWorkspace
	posX    float64
	posY    float64
	width   int
	height  int
	title   string
	client  *hypripc.Client
	appInfo *hyprpanelv1.AppInfo

	icon      *gtk.Image
	container *gtk.CenterBox
}

func (c *pagerClient) updateIcon() {
	var err error
	if c.appInfo == nil {
		c.appInfo, err = c.ws.pager.panel.host.FindApplication(c.client.Class)
		if err != nil || c.appInfo == nil {
			return
		}
	}

	if c.icon == nil {
		if c.width <= int(c.ws.pager.cfg.IconSize) || c.height <= int(c.ws.pager.cfg.IconSize) {
			return
		}
		if c.icon, err = createIcon(c.appInfo.Icon, int(c.ws.pager.cfg.IconSize), false, nil); err != nil {
			return
		}
		c.container.SetCenterWidget(&c.icon.Widget)
		return
	}

	if c.width <= int(c.ws.pager.cfg.IconSize) || c.height <= int(c.ws.pager.cfg.IconSize) {
		c.container.SetCenterWidget(&gtk.Widget{})
		icon := c.icon
		defer icon.Unref()
		c.icon = nil
	}
}

func (c *pagerClient) update(container *gtk.Fixed, posX, posY float64, width, height int, client *hypripc.Client) {
	if c.client != client {
		c.client = client
	}
	if c.client.Address == c.ws.pager.activeClient {
		if !c.container.HasCssClass(style.ActiveClass) {
			c.container.AddCssClass(style.ActiveClass)
		}
	} else if c.container.HasCssClass(style.ActiveClass) {
		c.container.RemoveCssClass(style.ActiveClass)
	}

	if c.title != c.client.Title {
		c.title = c.client.Title
		c.container.SetTooltipText(c.client.Title)
	}

	if c.width != width || c.height != height {
		c.width = width
		c.height = height
		c.updateIcon()
		c.container.SetSizeRequest(c.width, c.height)
	}

	if c.posX != posX || c.posY != posY {
		c.posX = posX
		c.posY = posY
		container.Move(&c.container.Widget, c.posX, c.posY)
	}
}

func (c *pagerClient) build(container *gtk.Fixed) {
	c.container = gtk.NewCenterBox()
	c.AddRef(c.container.Unref)
	c.container.AddCssClass(style.ClientClass)
	c.container.SetSizeRequest(c.width, c.height)
	c.container.SetMarginStart(1)
	c.container.SetMarginEnd(1)
	c.container.SetMarginTop(1)
	c.container.SetMarginBottom(1)

	c.updateIcon()
	c.update(container, c.posX, c.posY, c.width, c.height, c.client)

	container.Put(&c.container.Widget, c.posX, c.posY)
}

func (c *pagerClient) close(container *gtk.Fixed) {
	defer c.Unref()
	container.Remove(&c.container.Widget)
}

func newPagerClient(ws *pagerWorkspace, posX, posY float64, width, height int, client *hypripc.Client) *pagerClient {
	return &pagerClient{
		refTracker: newRefTracker(),
		ws:         ws,
		posX:       posX,
		posY:       posY,
		width:      width,
		height:     height,
		client:     client,
	}
}
