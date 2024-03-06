// Package gtk4layershell provides purego bindings to the gtk4-layer-shell library.
package gtk4layershell

import (
	"github.com/jwijenbergh/purego"
	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gtk"

	"fmt"
)

// Edge enum for screen edges.
type Edge int

const (
	// LayerShellEdgeLeft enum value
	LayerShellEdgeLeft Edge = iota
	// LayerShellEdgeRight enum value
	LayerShellEdgeRight
	// LayerShellEdgeTop enum value
	LayerShellEdgeTop
	// LayerShellEdgeBottom enum value
	LayerShellEdgeBottom
	// LayerShellEdgeEntryNumber should not be used except to get the number of entries
	LayerShellEdgeEntryNumber
)

func (e Edge) String() string {
	switch e {
	case LayerShellEdgeLeft:
		return `Left`
	case LayerShellEdgeRight:
		return `Right`
	case LayerShellEdgeTop:
		return `Top`
	case LayerShellEdgeBottom:
		return `Bottom`
	case LayerShellEdgeEntryNumber:
		return `EntryNumber`
	default:
		return fmt.Sprintf("Edge(%d)", e)
	}
}

// Layer enum
type Layer int

const (
	// LayerShellLayerBackground enum value
	LayerShellLayerBackground Layer = iota
	// LayerShellLayerBottom enum value
	LayerShellLayerBottom
	// LayerShellLayerTop enum value
	LayerShellLayerTop
	// LayerShellLayerOverlay enum value
	LayerShellLayerOverlay
	// LayerShellLayerEntryNumber should not be used except to get the number of entries
	LayerShellLayerEntryNumber
)

func (l Layer) String() string {
	switch l {
	case LayerShellLayerBackground:
		return `Background`
	case LayerShellLayerBottom:
		return `Bottom`
	case LayerShellLayerTop:
		return `Top`
	case LayerShellLayerOverlay:
		return `Overlay`
	case LayerShellLayerEntryNumber:
		return `EntryNumber`
	default:
		return fmt.Sprintf("Layer(%d)", l)
	}
}

// InitForWindow wraps gtk_layer_init_for_window
func InitForWindow(window *gtk.Window) {
	xInitForWindow(window.GoPointer())
}

var xInitForWindow func(uintptr)

// SetMonitor wraps gtk_layer_set_monitor
func SetMonitor(window *gtk.Window, monitor *gdk.Monitor) {
	xSetMonitor(window.GoPointer(), monitor.GoPointer())
}

var xSetMonitor func(uintptr, uintptr)

// AutoExclusiveZoneEnable wraps gtk_layer_auto_exclusive_zone_enable
func AutoExclusiveZoneEnable(window *gtk.Window) {
	xAutoExclusiveZoneEnable(window.GoPointer())
}

var xAutoExclusiveZoneEnable func(uintptr)

// SetAnchor wraps gtk_layer_set_anchor
func SetAnchor(window *gtk.Window, edge Edge, anchorToEdge bool) {
	xSetAnchor(window.GoPointer(), edge, anchorToEdge)
}

var xSetAnchor func(uintptr, Edge, bool)

// SetMargin wraps gtk_layer_set_margin
func SetMargin(window *gtk.Window, edge Edge, margin int) {
	xSetMargin(window.GoPointer(), edge, margin)
}

var xSetMargin func(uintptr, Edge, int)

// SetLayer wraps gtk_layer_set_layer
func SetLayer(window *gtk.Window, layer Layer) {
	xSetLayer(window.GoPointer(), layer)
}

var xSetLayer func(uintptr, Layer)

// SetNamespace wraps gtk_layer_set_namespace
func SetNamespace(window *gtk.Window, namespace string) {
	xSetNamespace(window.GoPointer(), namespace)
}

var xSetNamespace func(uintptr, string)

func puregoSafeRegister(fptr any, handle uintptr, name string) error {
	sym, err := purego.Dlsym(handle, name)
	if err != nil {
		return err
	}
	purego.RegisterFunc(fptr, sym)

	return nil
}

func init() {
	lib, err := purego.Dlopen(`libgtk4-layer-shell.so.0`, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(err)
	}

	if err := puregoSafeRegister(&xInitForWindow, lib, `gtk_layer_init_for_window`); err != nil {
		panic(err)
	}
	if err := puregoSafeRegister(&xSetMonitor, lib, `gtk_layer_set_monitor`); err != nil {
		panic(err)
	}
	if err := puregoSafeRegister(&xAutoExclusiveZoneEnable, lib, `gtk_layer_auto_exclusive_zone_enable`); err != nil {
		panic(err)
	}
	if err := puregoSafeRegister(&xSetAnchor, lib, `gtk_layer_set_anchor`); err != nil {
		panic(err)
	}
	if err := puregoSafeRegister(&xSetMargin, lib, `gtk_layer_set_margin`); err != nil {
		panic(err)
	}
	if err := puregoSafeRegister(&xSetLayer, lib, `gtk_layer_set_layer`); err != nil {
		panic(err)
	}
	if err := puregoSafeRegister(&xSetNamespace, lib, `gtk_layer_set_namespace`); err != nil {
		panic(err)
	}
}
