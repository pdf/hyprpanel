package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gdkpixbuf"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pdf/hyprpanel/internal/hypripc"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
)

type refTracker struct {
	mu   sync.Mutex
	refs []func()
}

func (r *refTracker) AddRef(f func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refs = append(r.refs, f)
}

func (r *refTracker) Unref() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ref := range r.refs {
		ref()
	}
}

func newRefTracker() *refTracker {
	return &refTracker{
		refs: make([]func(), 0),
	}
}

type callbackStore[T any] struct {
	mu   sync.RWMutex
	ptrs map[uintptr]*T
}

func (c *callbackStore[T]) Save(v T) uintptr {
	c.mu.Lock()
	defer c.mu.Unlock()
	val := &v
	ptr := uintptr(unsafe.Pointer(val))
	c.ptrs[ptr] = val

	return ptr
}

func (c *callbackStore[T]) Restore(ptr uintptr) (*T, error) {
	if ptr == 0 {
		return nil, errors.New(`nil pointer in callback restore`)
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.ptrs[ptr]
	if !ok {
		return nil, errors.New(`invalid pointer in callback restore`)
	}

	return v, nil
}

func (c *callbackStore[T]) Unref(ptr uintptr) {
	if ptr == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.ptrs, ptr)
}

var cbString = &callbackStore[string]{ptrs: make(map[uintptr]*string)}

func gdkMonitorFromHypr(monitor *hypripc.Monitor) (*gdk.Monitor, error) {
	disp := gdk.DisplayGetDefault()
	gdkMonitors := disp.GetMonitors()
	for i := uint(0); i < gdkMonitors.GetNItems(); i++ {
		p := gdkMonitors.GetItem(i)
		gmon := gdk.MonitorNewFromInternalPtr(p)
		gmon.GetConnector()
		if monitor.Name == gmon.GetConnector() {
			return gmon, nil
		}
	}

	return nil, errors.New(`monitor match not found`)
}

func pixbufFromSNIData(buf *eventv1.StatusNotifierValue_Pixmap, size int) (*gdkpixbuf.Pixbuf, error) {
	if len(buf.Data) == 0 ||
		len(buf.Data) != 4*int(buf.Width)*int(buf.Height) {
		return nil, errInvalidPixbufArray
	}

	// Convert ARGB to RGBA
	// TODO: Deal with endianness
	for i := 0; i < 4*int(buf.Width)*int(buf.Height); i += 4 {
		alpha := buf.Data[i]
		buf.Data[i] = buf.Data[i+1]
		buf.Data[i+1] = buf.Data[i+2]
		buf.Data[i+2] = buf.Data[i+3]
		buf.Data[i+3] = alpha
	}

	pixbuf := gdkpixbuf.NewPixbufFromBytes(glib.NewBytes((uintptr)(unsafe.Pointer(&buf.Data[0])), uint(len(buf.Data))), gdkpixbuf.GdkColorspaceRgbValue, true, 8, int(buf.Width), int(buf.Height), int(buf.Width)*4)
	scaled, err := pixbufScale(pixbuf, size)
	if err != nil {
		return nil, err
	}

	return scaled, nil
}

func pixbufFromNotificationData(buf *eventv1.NotificationValue_Pixmap, size int) (*gdkpixbuf.Pixbuf, error) {
	if len(buf.Data) == 0 ||
		int32(len(buf.Data)) != buf.Channels*buf.Width*buf.Height {
		return nil, errInvalidPixbufArray
	}

	pixbuf := gdkpixbuf.NewPixbufFromBytes(glib.NewBytes((uintptr)(unsafe.Pointer(&buf.Data[0])), uint(len(buf.Data))), gdkpixbuf.GdkColorspaceRgbValue, buf.HasAlpha, int(buf.BitsPerSample), int(buf.Width), int(buf.Height), int(buf.RowStride))
	if pixbuf == nil {
		return nil, errInvalidPixbufArray
	}
	scaled, err := pixbufScale(pixbuf, size)
	if err != nil {
		return nil, err
	}

	return scaled, nil
}

func pixbufScale(pixbuf *gdkpixbuf.Pixbuf, size int) (*gdkpixbuf.Pixbuf, error) {
	width := pixbuf.GetWidth()
	height := pixbuf.GetHeight()
	if (width > size || height > size) ||
		(width < size && height < size) {
		targetWidth, targetHeight := size, size
		if width > height {
			scale := float64(height) / float64(width)
			targetWidth = size
			targetHeight = int(math.Floor(scale * float64(size)))
		} else if height > width {
			scale := float64(width) / float64(height)
			targetHeight = size
			targetWidth = int(math.Floor(scale * float64(size)))
		}
		result := pixbuf.ScaleSimple(targetWidth, targetHeight, gdkpixbuf.GdkInterpBilinearValue)
		if result == nil {
			return nil, errors.New(`failed scaling pixbuf`)
		}

		return result, nil
	}

	return pixbuf, nil
}

func createIcon(icon string, size int, symbolic bool, fallbacks []string, searchPaths ...string) (*gtk.Image, error) {
	if strings.HasPrefix(icon, `/`) {
		pixbuf, err := gdkpixbuf.NewPixbufFromFileAtSize(icon, size, size)
		if err != nil {
			return nil, err
		}

		image := gtk.NewImageFromPixbuf(pixbuf)
		if image == nil {
			return nil, fmt.Errorf("could not convert icon pixbuf to image: %s", icon)
		}
		image.SetPixelSize(size)

		return image, nil
	}

	theme := gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault())
	if theme == nil {
		return nil, errors.New(`could not find default icon theme`)
	}
	for _, path := range searchPaths {
		if path != `` {
			theme.AddSearchPath(path)
		}
	}
	flags := gtk.IconLookupPreloadValue
	if symbolic {
		flags |= gtk.IconLookupForceSymbolicValue
	}
	iconInfo := theme.LookupIcon(icon, fallbacks, size, 1, gtk.TextDirLtrValue, flags)
	if iconInfo == nil {
		return nil, fmt.Errorf("icon not found in theme (%s): %s", theme.GetThemeName(), icon)
	}
	imageWidget := gtk.NewImageFromPaintable(iconInfo)
	if imageWidget == nil {
		return nil, fmt.Errorf("could not convert icon to image: %s", icon)
	}
	var image gtk.Image
	imageWidget.Widget.Cast(&image)
	image.SetPixelSize(size)
	return &image, nil
}
