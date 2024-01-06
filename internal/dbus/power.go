package dbus

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/hashicorp/go-hclog"
	"github.com/iancoleman/strcase"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	powerHudId = `power`

	powerDisplayDevice = `DisplayDevice`
)

type power struct {
	sync.RWMutex
	conn *dbus.Conn
	log  hclog.Logger
	cfg  *configv1.Config_DBUS_Power

	displayDevicePath dbus.ObjectPath

	cachePower map[string]*eventv1.PowerChangeValue

	eventCh chan *eventv1.Event
	signals chan *dbus.Signal
	readyCh chan struct{}
	quitCh  chan struct{}
}

func (b *power) updatePower(objPath dbus.ObjectPath, props map[string]dbus.Variant) error {
	name := filepath.Base(string(objPath))
	if name == powerDisplayDevice {
		name = eventv1.PowerDefaultId
	}

	powerValue := &eventv1.PowerChangeValue{
		Id: name,
	}
	lastValue, ok := b.cachePower[name]
	if ok {
		proto.Merge(powerValue, lastValue)
	} else if !ok || props == nil {
		lastValue = &eventv1.PowerChangeValue{}
		busObj := b.conn.Object(fdoUPowerName, objPath)
		call := busObj.Call(fdoPropertiesMethodGetAll, 0, fdoUPowerDeviceName)
		if call.Err != nil {
			return fmt.Errorf("failed getting power device properties: %w", call.Err)
		}

		props = make(map[string]dbus.Variant)
		if err := call.Store(&props); err != nil {
			return err
		}
	}

	hudUpdate := false
	for k, v := range props {
		switch k {
		case fdoUPowerDevicePropertyVendor, strcase.ToKebab(fdoUPowerDevicePropertyVendor):
			var val string
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.Vendor = val
		case fdoUPowerDevicePropertyModel, strcase.ToKebab(fdoUPowerDevicePropertyModel):
			var val string
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.Model = val
		case fdoUPowerDevicePropertyType, strcase.ToKebab(fdoUPowerDevicePropertyType):
			var val eventv1.PowerType
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.Type = val
		case fdoUPowerDevicePropertyPowerSupply, strcase.ToKebab(fdoUPowerDevicePropertyPowerSupply):
			var val bool
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.PowerSupply = val
		case fdoUPowerDevicePropertyOnline, strcase.ToKebab(fdoUPowerDevicePropertyOnline):
			var val bool
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.Online = val
		case fdoUPowerDevicePropertyTimeToEmpty, strcase.ToKebab(fdoUPowerDevicePropertyTimeToEmpty):
			var val int64
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.TimeToEmpty = durationpb.New(time.Duration(val) * time.Second)
		case fdoUPowerDevicePropertyTimeToFull, strcase.ToKebab(fdoUPowerDevicePropertyTimeToFull):
			var val int64
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.TimeToFull = durationpb.New(time.Duration(val) * time.Second)
		case fdoUPowerDevicePropertyPercentage, strcase.ToKebab(fdoUPowerDevicePropertyPercentage):
			var val float64
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.Percentage = uint32(math.Round(val))
			if powerValue.Percentage <= b.cfg.LowPercent && powerValue.Percentage != lastValue.Percentage {
				hudUpdate = true
			}
		case fdoUPowerDevicePropertyState, strcase.ToKebab(fdoUPowerDevicePropertyState):
			var val eventv1.PowerState
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.State = val
			if powerValue.State != lastValue.State {
				hudUpdate = true
			}
		case fdoUPowerDevicePropertyIconName, strcase.ToKebab(fdoUPowerDevicePropertyIconName):
			var val string
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.Icon = val
			if powerValue.Icon != lastValue.Icon {
				hudUpdate = true
			}
		case fdoUPowerDevicePropertyEnergy, strcase.ToKebab(fdoUPowerDevicePropertyEnergy):
			var val float64
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.Energy = val
		case fdoUPowerDevicePropertyEnergyEmpty, strcase.ToKebab(fdoUPowerDevicePropertyEnergyEmpty):
			var val float64
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.EnergyEmpty = val
		case fdoUPowerDevicePropertyEnergyFull, strcase.ToKebab(fdoUPowerDevicePropertyEnergyFull):
			var val float64
			if err := v.Store(&val); err != nil {
				return err
			}
			powerValue.EnergyFull = val
		}
	}

	b.cachePower[powerValue.Id] = powerValue

	powerData, err := anypb.New(powerValue)
	if err != nil {
		return fmt.Errorf(`failed encodiung event data for power dev (%s): %w`, name, err)
	}

	b.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_DBUS_POWER_CHANGE,
		Data: powerData,
	}

	if !b.cfg.HudNotifications || !hudUpdate || powerValue.Type != eventv1.PowerType_POWER_TYPE_BATTERY {
		return nil
	}

	select {
	case <-b.readyCh:
	default:
		return nil
	}

	var body strings.Builder
	if powerValue.Vendor != `` {
		body.WriteString(powerValue.Vendor)
		if powerValue.Model != `` {
			body.WriteString(` `)
			body.WriteString(powerValue.Model)
		}
		body.WriteString("\r\r")
	}

	switch powerValue.State {
	case eventv1.PowerState_POWER_STATE_CHARGING:
		body.WriteString(`Charging`)
	case eventv1.PowerState_POWER_STATE_DISCHARGING:
		body.WriteString(`Discharging`)
	case eventv1.PowerState_POWER_STATE_EMPTY:
		body.WriteString(`Empty!`)
	case eventv1.PowerState_POWER_STATE_FULLY_CHARGED:
		body.WriteString(`Fully Charged`)
	case eventv1.PowerState_POWER_STATE_PENDING_CHARGE:
		body.WriteString(`Pending Charge`)
	case eventv1.PowerState_POWER_STATE_PENDING_DISCHARGE:
		body.WriteString(`Pending Discharge`)
	}

	hudValue := &eventv1.HudNotificationValue{
		Id:           powerHudId,
		Icon:         powerValue.Icon,
		IconSymbolic: true,
		Title:        powerValue.Id,
		Body:         body.String(),
		Percent:      (powerValue.Energy - powerValue.EnergyEmpty) / (powerValue.EnergyFull - powerValue.EnergyEmpty),
	}

	hudData, err := anypb.New(hudValue)
	if err != nil {
		return err
	}

	b.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_HUD_NOTIFY,
		Data: hudData,
	}

	return nil
}

func (b *power) init() error {
	busObj := b.conn.Object(fdoUPowerName, fdoUPowerPath)
	call := busObj.Call(fdoUPowerMethodGetDisplayDevice, 0)
	if call.Err != nil {
		return fmt.Errorf("failed getting default power device: %w", call.Err)
	}

	if err := call.Store(&b.displayDevicePath); err != nil {
		return err
	}

	deviceObj := b.conn.Object(fdoUPowerDeviceName, b.displayDevicePath)
	if err := deviceObj.AddMatchSignal(fdoPropertiesName, fdoPropertiesMemberPropertiesChanged).Err; err != nil {
		return err
	}

	if err := b.updatePower(b.displayDevicePath, nil); err != nil {
		return err
	}

	close(b.readyCh)

	go b.watch()

	return nil
}

func (b *power) watch() {
	for {
		select {
		case <-b.quitCh:
			return
		default:
			select {
			case <-b.quitCh:
				return
			case sig, ok := <-b.signals:
				if !ok {
					return
				}
				switch sig.Name {
				case fdoPropertiesSignalPropertiesChanged:
					if sig.Path != b.displayDevicePath {
						continue
					}
					if len(sig.Body) != 3 {
						b.log.Warn(`Failed parsing DBUS PropertiesChanged body`, `body`, sig.Body)
						continue
					}

					kind, ok := sig.Body[0].(string)
					if !ok {
						b.log.Warn(`Failed asserting DBUS PropertiesChanged body kind`, `kind`, sig.Body[0])
						continue
					}
					if kind != fdoUPowerDeviceName {
						continue
					}

					properties, ok := sig.Body[1].(map[string]dbus.Variant)
					if !ok {
						b.log.Warn(`Failed asserting DBUS PropertiesChanged body properties`, `properties`, sig.Body[1])
						continue
					}
					if len(properties) == 0 {
						continue
					}

					if err := b.updatePower(sig.Path, properties); err != nil {
						b.log.Warn(`Failed polling power`, `path`, sig.Path, `err`, err)
						continue
					}
				}
			}
		}
	}
}

func newPower(conn *dbus.Conn, logger hclog.Logger, eventCh chan *eventv1.Event, cfg *configv1.Config_DBUS_Power) (*power, error) {
	s := &power{
		conn:       conn,
		log:        logger,
		cfg:        cfg,
		cachePower: make(map[string]*eventv1.PowerChangeValue),
		eventCh:    eventCh,
		signals:    make(chan *dbus.Signal),
		readyCh:    make(chan struct{}),
		quitCh:     make(chan struct{}),
	}

	s.conn.Signal(s.signals)

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
