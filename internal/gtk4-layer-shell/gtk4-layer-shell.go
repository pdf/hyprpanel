package gtk4layershell

import (
	"github.com/jwijenbergh/purego"
	"github.com/jwijenbergh/puregotk/v4/gdk"
	"github.com/jwijenbergh/puregotk/v4/gtk"

	"fmt"
)

type Edge int

const (
	LayerShellEdgeLeft Edge = iota
	LayerShellEdgeRight
	LayerShellEdgeTop
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

type Layer int

const (
	LayerShellLayerBackground Layer = iota
	LayerShellLayerBottom
	LayerShellLayerTop
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

func InitForWindow(window *gtk.Window) {
	xInitForWindow(window.GoPointer())
}

var xInitForWindow func(uintptr)

func SetMonitor(window *gtk.Window, monitor *gdk.Monitor) {
	xSetMonitor(window.GoPointer(), monitor.GoPointer())
}

var xSetMonitor func(uintptr, uintptr)

func AutoExclusiveZoneEnable(window *gtk.Window) {
	xAutoExclusiveZoneEnable(window.GoPointer())
}

var xAutoExclusiveZoneEnable func(uintptr)

func SetAnchor(window *gtk.Window, edge Edge, anchorToEdge bool) {
	xSetAnchor(window.GoPointer(), edge, anchorToEdge)
}

var xSetAnchor func(uintptr, Edge, bool)

func SetMargin(window *gtk.Window, edge Edge, margin int) {
	xSetMargin(window.GoPointer(), edge, margin)
}

var xSetMargin func(uintptr, Edge, int)

func SetLayer(window *gtk.Window, layer Layer) {
	xSetLayer(window.GoPointer(), layer)
}

var xSetLayer func(uintptr, Layer)

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

	puregoSafeRegister(&xInitForWindow, lib, `gtk_layer_init_for_window`)
	puregoSafeRegister(&xSetMonitor, lib, `gtk_layer_set_monitor`)
	puregoSafeRegister(&xAutoExclusiveZoneEnable, lib, `gtk_layer_auto_exclusive_zone_enable`)
	puregoSafeRegister(&xSetAnchor, lib, `gtk_layer_set_anchor`)
	puregoSafeRegister(&xSetMargin, lib, `gtk_layer_set_margin`)
	puregoSafeRegister(&xSetLayer, lib, `gtk_layer_set_layer`)
	puregoSafeRegister(&xSetNamespace, lib, `gtk_layer_set_namespace`)
}
