package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gdkpixbuf"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/pdf/hyprpanel/internal/hypripc"
	"github.com/pdf/hyprpanel/internal/panelplugin"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	hyprpanelv1 "github.com/pdf/hyprpanel/proto/hyprpanel/v1"
	"github.com/pdf/hyprpanel/style"
)

const tooltipDebounceTime = 500 * time.Millisecond

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
	if len(buf.Data) == 0 || int32(len(buf.Data)) != buf.Channels*buf.Width*buf.Height {
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

func pixbufFromNRGBA(buf *hyprpanelv1.ImageNRGBA) (*gdkpixbuf.Pixbuf, error) {
	if len(buf.Pixels) == 0 {
		return nil, errInvalidPixbufArray
	}

	pixbuf := gdkpixbuf.NewPixbufFromBytes(glib.NewBytes((uintptr)(unsafe.Pointer(&buf.Pixels[0])), uint(len(buf.Pixels))), gdkpixbuf.GdkColorspaceRgbValue, true, 8, int(buf.Width), int(buf.Height), int(buf.Stride))
	if pixbuf == nil {
		return nil, errInvalidPixbufArray
	}

	return pixbuf, nil
}

func pixbufScale(pixbuf *gdkpixbuf.Pixbuf, size int) (*gdkpixbuf.Pixbuf, error) {
	width := pixbuf.GetWidth()
	height := pixbuf.GetHeight()
	if (width == size && height <= size) || (height == size && width <= size) {
		return pixbuf, nil
	}
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

func unrefCallback(fnPtr any) {
	if err := glib.UnrefCallback(fnPtr); err != nil {
		log.Warn(`UnrefCallback failed`, `err`, err)
	}
}

type tooltipPreviewer interface {
	clientAddress() string
	clientTitle() string
	shouldPreview() bool
	host() panelplugin.Host
}

func tooltipPreview(target tooltipPreviewer, width, height int) func(widget gtk.Widget, x int, y int, keyboardMod bool, tooltipPtr uintptr) bool {
	var lastTooltipTime time.Time
	var lastTooltip *gtk.Widget

	return func(widget gtk.Widget, x int, y int, keyboardMod bool, tooltipPtr uintptr) bool {
		tooltip := gtk.TooltipNewFromInternalPtr(tooltipPtr)

		// Debounce tooltip rendering to avoid excessive updates
		if !lastTooltipTime.IsZero() && time.Since(lastTooltipTime) < tooltipDebounceTime && lastTooltip != nil {
			tooltip.SetCustom(lastTooltip)
			return true
		}
		lastTooltipTime = time.Now()
		time.AfterFunc(tooltipDebounceTime, func() {
			if lastTooltip == nil {
				return
			}
			lastTooltip.Unref()
			lastTooltip = nil
		})

		container := gtk.NewBox(gtk.OrientationVerticalValue, 0)
		tooltip.SetCustom(&container.Widget)
		lastTooltip = &container.Widget

		title := gtk.NewLabel(target.clientTitle())
		container.Append(&title.Widget)
		title.Unref()

		if !target.shouldPreview() {
			return true
		}

		addr, err := strconv.ParseUint(target.clientAddress(), 0, 64)
		if err != nil {
			log.Warn(`failed to parse client address`, `err`, err)
			return true
		}
		img, err := target.host().CaptureFrame(addr, int32(width), int32(height))
		if err != nil {
			log.Warn(`failed to capture frame`, `err`, err)
			return true
		}

		pixbuf, err := pixbufFromNRGBA(img)
		if err != nil {
			log.Warn(`failed to create pixbuf from ImageNRGBA`, `err`, err)
			return true
		}
		image := gtk.NewPictureForPixbuf(pixbuf)
		image.AddCssClass(style.TooltipImageClass)
		image.SetSizeRequest(width, height)
		container.Append(&image.Widget)
		image.Unref()

		return true
	}
}
