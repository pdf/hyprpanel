package main

import (
	"math"
	"strconv"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/jwijenbergh/puregotk/v4/pango"
	"github.com/pdf/hyprpanel/internal/hypripc"
	"github.com/pdf/hyprpanel/style"
)

type pagerWorkspace struct {
	*refTracker
	pager   *pager
	id      int
	name    string
	pinned  bool
	width   int
	height  int
	live    bool
	clients map[string]*pagerClient

	container *gtk.Box
	inner     *gtk.Fixed
	label     *gtk.Label
}

func (w *pagerWorkspace) rename(name string) {
	if w.name != name {
		w.name = name
		w.label.SetText(w.name)
	}
}

func (w *pagerWorkspace) updateClient(hyprclient *hypripc.Client) {
	width := int(math.Floor(float64(hyprclient.Size[0]) * w.pager.scale * w.pager.panel.currentMonitor.Scale))
	height := int(math.Floor(float64(hyprclient.Size[1]) * w.pager.scale * w.pager.panel.currentMonitor.Scale))
	posX := math.Floor(float64(hyprclient.At[0]) * w.pager.scale * w.pager.panel.currentMonitor.Scale)
	posY := math.Floor(float64(hyprclient.At[1]) * w.pager.scale * w.pager.panel.currentMonitor.Scale)

	if widthDelta := w.width - int(posX) - width; widthDelta < 0 {
		width += widthDelta
	}
	if heightDelta := w.height - int(posY) - height; heightDelta < 0 {
		height += heightDelta
	}

	// margins
	width -= 2
	height -= 2

	if hyprclient.Class == `` {
		if hyprclient.InitialClass != `` {
			hyprclient.Class = hyprclient.InitialClass
		} else {
			hyprclient.Class = hyprclient.InitialTitle
		}
	}

	if client, ok := w.clients[hyprclient.Address]; ok {
		if (hyprclient.Hidden && hyprclient.Hidden != client.client.Hidden) ||
			(!hyprclient.Mapped && hyprclient.Mapped != client.client.Mapped) ||
			width < 2 || height < 2 {
			w.deleteClient(hyprclient.Address)
			return
		}

		client.update(w.inner, posX, posY, width, height, hyprclient)

		return
	}

	if hyprclient.Hidden || !hyprclient.Mapped || width < 2 || height < 2 {
		return
	}

	client := newPagerClient(w, posX, posY, width, height, hyprclient)
	client.build(w.inner)
	w.clients[hyprclient.Address] = client
}

func (w *pagerWorkspace) deleteClient(addr string) {
	client, ok := w.clients[addr]
	if !ok {
		return
	}

	client.close(w.inner)
	delete(w.clients, addr)
}

func (w *pagerWorkspace) setActive(live bool, active bool) {
	w.live = live
	if live {
		if !w.container.HasCssClass(style.LiveClass) {
			w.container.AddCssClass(style.LiveClass)
		}
	} else if w.container.HasCssClass(style.LiveClass) {
		w.container.RemoveCssClass(style.LiveClass)
	}
	if active {
		if !w.container.HasCssClass(style.ActiveClass) {
			w.container.AddCssClass(style.ActiveClass)
		}
	} else if w.container.HasCssClass(style.ActiveClass) {
		w.container.RemoveCssClass(style.ActiveClass)
	}
}

func (w *pagerWorkspace) build(container *gtk.Box) error {
	w.container = gtk.NewBox(gtk.OrientationVerticalValue, 0)
	w.AddRef(w.container.Unref)
	w.inner = gtk.NewFixed()
	w.AddRef(w.inner.Unref)

	if w.pager.panel.orientation == gtk.OrientationHorizontalValue {
		w.width = int(math.Floor(w.pager.scale*float64(w.pager.panel.currentMonitor.Width))) - 2
		w.height = int(math.Min(math.Floor(w.pager.scale*float64(w.pager.panel.currentMonitor.Height-int(w.pager.panel.cfg.Size))), float64(w.pager.panel.cfg.Size-2))) - 2
	} else {
		w.width = int(math.Min(math.Floor(w.pager.scale*float64(w.pager.panel.currentMonitor.Width-int(w.pager.panel.cfg.Size))), float64(w.pager.panel.cfg.Size-2))) - 2
		w.height = int(math.Floor(w.pager.scale*float64(w.pager.panel.currentMonitor.Height))) - 2
	}

	w.container.AddCssClass(style.WorkspaceClass)
	w.container.SetMarginStart(1)
	w.container.SetMarginEnd(1)
	w.container.SetMarginTop(1)
	w.container.SetMarginBottom(1)

	clickCb := func(ctrl gtk.GestureClick, _ int, _, _ float64) {
		switch ctrl.GetCurrentButton() {
		case uint(gdk.BUTTON_PRIMARY):
			if err := w.pager.panel.hypr.Dispatch(hypripc.DispatchWorkspace, strconv.Itoa(int(w.id))); err != nil {
				log.Warn(`Switch workspace failed`, `module`, style.PagerID, `err`, err)
			}
		}
	}
	w.AddRef(func() {
		unrefCallback(&clickCb)
	})
	clickController := gtk.NewGestureClick()
	clickController.ConnectReleased(&clickCb)
	w.container.AddController(&clickController.EventController)

	w.inner = gtk.NewFixed()
	w.container.Append(&w.inner.Widget)

	if w.pager.cfg.EnableWorkspaceNames {
		w.label = gtk.NewLabel(w.name)
		w.label.SetWrap(false)
		w.label.SetSingleLineMode(true)
		w.label.SetEllipsize(pango.EllipsizeEndValue)
		w.label.SetHalign(gtk.AlignCenterValue)
		w.label.SetValign(gtk.AlignCenterValue)
		w.label.SetHexpand(true)
		w.label.AddCssClass(style.WorkspaceLabelClass)
		w.container.Append(&w.label.Widget)
		w.AddRef(w.label.Unref)
	}

	w.container.SetSizeRequest(w.width, w.height)
	w.inner.SetSizeRequest(w.width, w.height)

	container.Append(&w.container.Widget)

	return nil
}

func (w *pagerWorkspace) close(container *gtk.Box) error {
	if w.pinned {
		return errPinned
	}
	defer w.Unref()
	for _, client := range w.clients {
		client.close(w.inner)
	}
	container.Remove(&w.container.Widget)

	return nil
}

func newPagerWorkspace(p *pager, id int, name string, pinned bool) *pagerWorkspace {
	return &pagerWorkspace{
		refTracker: newRefTracker(),
		pager:      p,
		id:         id,
		name:       name,
		pinned:     pinned,
		clients:    make(map[string]*pagerClient),
	}
}
