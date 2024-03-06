package dbus

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/hashicorp/go-hclog"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	brightnessHudID = `brightness`

	brightnessSysfsDevice = `device`

	brightnessBase    = `/sys/class/backlight`
	brightnessNode    = `brightness`
	brightnessMaxNode = `max_brightness`
)

type brightness struct {
	sync.RWMutex
	conn *dbus.Conn
	log  hclog.Logger
	cfg  *configv1.Config_DBUS_Brightness

	cacheBrightness map[string]*eventv1.BrightnessChangeValue

	eventCh chan *eventv1.Event
	signals chan *dbus.Signal
	readyCh chan struct{}
	quitCh  chan struct{}
}

func (b *brightness) Adjust(devName string, direction eventv1.Direction) error {
	if devName != `` {
		return b.adjust(filepath.Join(brightnessBase, devName), direction)
	}

	targets, err := os.ReadDir(brightnessBase)
	if err != nil {
		return err
	}
	for _, target := range targets {
		path := filepath.Join(brightnessBase, target.Name())
		if target.Type()&fs.ModeSymlink == fs.ModeSymlink {
			path, err = os.Readlink(filepath.Join(brightnessBase, target.Name()))
			if err != nil {
				return err
			}
			path = filepath.Join(brightnessBase, path)
		} else if !target.IsDir() {
			continue
		}

		if err := b.adjust(path, direction); err != nil {
			return err
		}
	}

	return nil
}

func (b *brightness) adjust(path string, direction eventv1.Direction) error {
	maxB, err := os.ReadFile(filepath.Join(path, brightnessMaxNode))
	if err != nil {
		return err
	}
	curB, err := os.ReadFile(filepath.Join(path, brightnessNode))
	if err != nil {
		return err
	}

	max, err := strconv.Atoi(strings.TrimSuffix(string(maxB), "\n"))
	if err != nil {
		return err
	}
	cur, err := strconv.Atoi(strings.TrimSuffix(string(curB), "\n"))
	if err != nil {
		return err
	}

	if direction == eventv1.Direction_DIRECTION_UP {
		if cur >= max {
			return nil
		}
		cur += max / 100 * int(b.cfg.AdjustStepPercent)
		if cur > max {
			cur = max
		}
	} else {
		if cur <= int(b.cfg.MinBrightness) {
			return nil
		}
		cur -= max / 100 * int(b.cfg.AdjustStepPercent)
		if cur < int(b.cfg.MinBrightness) {
			cur = int(b.cfg.MinBrightness)
		}
	}

	if !b.cfg.EnableLogind {
		return os.WriteFile(filepath.Join(path, brightnessNode), []byte(strconv.Itoa(cur)+"\n"), 0664)
	}

	obj := b.conn.Object(fdoLogindName, fdoLogindSessionPath)
	return obj.Call(fdoLogindSessionMethodSetBrightness, 0, `backlight`, filepath.Base(path), uint32(cur)).Err
}

func (b *brightness) pollBrightness(path string) error {
	maxB, err := os.ReadFile(filepath.Join(path, brightnessMaxNode))
	if err != nil {
		return err
	}
	curB, err := os.ReadFile(filepath.Join(path, brightnessNode))
	if err != nil {
		return err
	}

	max, err := strconv.Atoi(strings.TrimSuffix(string(maxB), "\n"))
	if err != nil {
		return err
	}
	cur, err := strconv.Atoi(strings.TrimSuffix(string(curB), "\n"))
	if err != nil {
		return err
	}

	dev, err := os.Readlink(filepath.Join(path, brightnessSysfsDevice))
	if err != nil {
		return err
	}

	brightnessValue := &eventv1.BrightnessChangeValue{
		Id:            filepath.Base(path),
		Name:          filepath.Base(dev),
		Brightness:    int32(cur),
		BrightnessMax: int32(max),
	}

	if v, ok := b.cacheBrightness[brightnessValue.Id]; ok {
		if v.Brightness == brightnessValue.Brightness && v.BrightnessMax == brightnessValue.Brightness {
			return nil
		}
	}

	b.cacheBrightness[brightnessValue.Id] = brightnessValue

	brightnessData, err := anypb.New(brightnessValue)
	if err != nil {
		return fmt.Errorf(`failed encodiung event data for brightness dev (%s): %w`, dev, err)
	}

	b.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_DBUS_BRIGHTNESS_CHANGE,
		Data: brightnessData,
	}

	if !b.cfg.HudNotifications {
		return nil
	}

	select {
	case <-b.readyCh:
	default:
		return nil
	}

	percent := math.Round(float64(brightnessValue.Brightness)/float64(brightnessValue.BrightnessMax)*100) / 100
	icon := `display-brightness-off`
	switch {
	case percent >= 1:
		icon = `display-brightness-high`
	case percent >= 0.5:
		icon = `display-brightness-medium`
	case cur > int(b.cfg.MinBrightness):
		icon = `display-brightness-low`
	}

	hudValue := &eventv1.HudNotificationValue{
		Id:           brightnessHudID,
		Icon:         icon,
		IconSymbolic: true,
		Title:        brightnessValue.Id,
		Body:         brightnessValue.Name,
		Percent:      percent,
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

func (b *brightness) init() error {
	targets, err := os.ReadDir(brightnessBase)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if target.Type()&fs.ModeSymlink != fs.ModeSymlink {
			continue
		}
		path, err := os.Readlink(filepath.Join(brightnessBase, target.Name()))
		if err != nil {
			return err
		}
		path = filepath.Join(brightnessBase, path)
		if err := b.pollBrightness(path); err != nil {
			b.log.Error(`Failed retrieving brightness`, `err`, err)
		}

		objectPath, err := systemdUnitToObjectPath(path + `.device`)
		if err != nil {
			b.log.Warn(`Failed encoding brightness unit name`, `path`, path, `err`, err)
		}

		busObj := b.conn.Object(fdoSystemdName, fdoSystemdUnitPath+objectPath)
		if busObj.AddMatchSignal(fdoPropertiesName, fdoPropertiesMemberPropertiesChanged).Err != nil {
			return err
		}
	}

	close(b.readyCh)

	go b.watch()

	return nil
}

func (b *brightness) watch() {
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
					if len(sig.Body) != 3 {
						b.log.Warn(`Failed parsing DBUS PropertiesChanged body`, `body`, sig.Body)
						continue
					}

					kind, ok := sig.Body[0].(string)
					if !ok {
						b.log.Warn(`Failed asserting DBUS PropertiesChanged body kind`, `kind`, sig.Body[0])
						continue
					}
					if kind != fdoSystemdDeviceName {
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

					pathVar, ok := properties[`SysFSPath`]
					if !ok {
						continue
					}
					var path string
					if err := pathVar.Store(&path); err != nil {
						b.log.Warn(`Failed parsing SysFSPath`, `pathVar`, pathVar, `err`, err)
						continue
					}
					if path == `` {
						continue
					}
					if err := b.pollBrightness(path); err != nil {
						b.log.Warn(`Failed polling brightness`, `path`, path, `err`, err)
						continue
					}
				}
			}
		}
	}
}

func newBrightness(conn *dbus.Conn, logger hclog.Logger, eventCh chan *eventv1.Event, cfg *configv1.Config_DBUS_Brightness) (*brightness, error) {
	s := &brightness{
		conn:            conn,
		log:             logger,
		cfg:             cfg,
		cacheBrightness: make(map[string]*eventv1.BrightnessChangeValue),
		eventCh:         eventCh,
		signals:         make(chan *dbus.Signal),
		readyCh:         make(chan struct{}),
		quitCh:          make(chan struct{}),
	}

	s.conn.Signal(s.signals)

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}
