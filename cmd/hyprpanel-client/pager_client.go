package main

import (
	"fmt"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pdf/hyprpanel/internal/hypripc"
	modulev1 "github.com/pdf/hyprpanel/proto/hyprpanel/module/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/pdf/hyprpanel/style"
)

type pagerClient struct {
	*refTracker
	*api
	cfg     *modulev1.Pager
	active  bool
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
		c.appInfo, err = c.host.FindApplication(c.client.Class)
		if err != nil || c.appInfo == nil {
			return
		}
	}

	if c.icon == nil {
		if c.width <= int(c.cfg.IconSize) || c.height <= int(c.cfg.IconSize) {
			return
		}
		if c.icon, err = createIcon(c.appInfo.Icon, int(c.cfg.IconSize), false, nil); err != nil {
			return
		}
		c.container.SetCenterWidget(&c.icon.Widget)
		return
	}

	if c.width <= int(c.cfg.IconSize) || c.height <= int(c.cfg.IconSize) {
		c.container.SetCenterWidget(&gtk.Widget{})
		icon := c.icon
		defer icon.Unref()
		c.icon = nil
	}
}

func (c *pagerClient) update(container *gtk.Fixed, posX, posY float64, width, height int, client *hypripc.Client, active bool) {
	if c.client != client {
		c.client = client
	}
	c.active = active
	if c.active {
		if !c.container.HasCssClass(style.ActiveClass) {
			c.container.AddCssClass(style.ActiveClass)
		}
	} else if c.container.HasCssClass(style.ActiveClass) {
		c.container.RemoveCssClass(style.ActiveClass)
	}

	if c.title != c.client.Title {
		c.title = c.client.Title
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

func (c *pagerClient) clientTitle() string {
	return c.title
}

func (c *pagerClient) clientSubtitle() string {
	mem, err := memKb(int(c.client.Pid))
	if err != nil {
		return ``
	}
	return fmt.Sprintf("Memory: %.1f MB", float64(mem)/1024.0)
}

func (c *pagerClient) clientAddress() string {
	return c.client.Address
}

func (c *pagerClient) shouldPreview() bool {
	return c.client != nil
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
	c.container.SetHasTooltip(true)

	previewHeight := int(c.cfg.PreviewWidth * 9 / 16)
	tooltipCb := tooltipPreview(c, int(c.cfg.PreviewWidth), previewHeight)
	c.AddRef(func() { unrefCallback(&tooltipCb) })
	c.container.ConnectQueryTooltip(&tooltipCb)

	dragSource := gtk.NewDragSource()
	dragPrepCb := func(_ gtk.DragSource, _, _ float64) gdk.ContentProvider {
		val := gobject.Value{GType: gobject.TypeStringVal}
		val.SetString(c.client.Address)
		return *gdk.NewContentProviderForValue(&val)
	}
	c.AddRef(func() { unrefCallback(&dragPrepCb) })
	dragSource.ConnectPrepare(&dragPrepCb)
	dragBeginCb := func(_ gtk.DragSource, _ uintptr) {
		preview := gtk.NewWidgetPaintable(&c.container.Widget)
		// hotX/hotY don't work here, apparently it's meant to be fixed in GTK, maybe Hyprland bug?
		// https://gitlab.gnome.org/GNOME/gtk/-/issues/2341
		// https://github.com/hyprwm/Hyprland/issues/9564
		dragSource.SetIcon(preview, preview.GetIntrinsicWidth()/2, preview.GetIntrinsicHeight()/2)
		preview.Unref()
	}
	dragSource.ConnectDragBegin(&dragBeginCb)
	c.container.AddController(&dragSource.EventController)

	c.updateIcon()
	c.update(container, c.posX, c.posY, c.width, c.height, c.client, c.active)

	container.Put(&c.container.Widget, c.posX, c.posY)
}

func (c *pagerClient) close(container *gtk.Fixed) {
	defer c.Unref()
	container.Remove(&c.container.Widget)
}

func newPagerClient(cfg *modulev1.Pager, a *api, posX, posY float64, width, height int, client *hypripc.Client, active bool) *pagerClient {
	return &pagerClient{
		refTracker: newRefTracker(),
		api:        a,
		cfg:        cfg,
		active:     active,
		posX:       posX,
		posY:       posY,
		width:      width,
		height:     height,
		client:     client,
	}
}
