// Package audio provides a high level pulseaudio API.
package audio

import (
	"fmt"
	"math"
	"net"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/jfreymuth/pulse/proto"
	configv1 "github.com/pdf/hyprpanel/proto/hyprpanel/config/v1"
	eventv1 "github.com/pdf/hyprpanel/proto/hyprpanel/event/v1"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	pulseClientName    = `hyprpanel`
	pulseIconName      = `audio-x-generic`
	hudID              = `audio`
	exceedVolumeFactor = 1.3
)

var (
	volumeMax = int32(math.Floor(float64(proto.VolumeNorm) * exceedVolumeFactor))
)

// Client for pulseaudio.
type Client struct {
	cfg   *configv1.Config_Audio
	log   hclog.Logger
	proto *proto.Client
	conn  net.Conn

	cacheSink   map[string]*eventv1.AudioSinkChangeValue
	cacheSource map[string]*eventv1.AudioSourceChangeValue

	eventCh  chan *eventv1.Event
	sinkCh   chan uint32
	sourceCh chan uint32
	readyCh  chan struct{}
	quitCh   chan struct{}
}

// SinkVolumeAdjust increases or decreases a sink's volume.
func (c *Client) SinkVolumeAdjust(sinkName string, direction eventv1.Direction) error {
	info := proto.GetSinkInfoReply{}
	if err := c.proto.Request(&proto.GetSinkInfo{SinkIndex: proto.Undefined, SinkName: sinkName}, &info); err != nil {
		return err
	}

	for i, v := range info.ChannelVolumes {
		vol := int32(v)
		if direction == eventv1.Direction_DIRECTION_UP {
			if c.cfg.VolumeExceedMaximum && vol >= volumeMax {
				return nil
			} else if !c.cfg.VolumeExceedMaximum && vol >= int32(proto.VolumeNorm) {
				return nil
			}

			vol += int32(proto.VolumeNorm) / 100 * int32(c.cfg.VolumeStepPercent)
			if c.cfg.VolumeExceedMaximum && vol > volumeMax {
				vol = volumeMax
			} else if !c.cfg.VolumeExceedMaximum && vol > int32(proto.VolumeNorm) {
				vol = int32(proto.VolumeNorm)
			}
		} else {
			if vol <= 0 {
				return nil
			}

			vol -= int32(proto.VolumeNorm) / 100 * int32(c.cfg.VolumeStepPercent)
			if vol < 0 {
				vol = 0
			}
		}

		info.ChannelVolumes[i] = uint32(vol)
	}

	if info.Mute && direction == eventv1.Direction_DIRECTION_UP {
		if err := c.proto.Request(&proto.SetSinkMute{SinkIndex: proto.Undefined, SinkName: sinkName, Mute: false}, nil); err != nil {
			return err
		}
	}

	if err := c.proto.Request(&proto.SetSinkVolume{SinkIndex: proto.Undefined, SinkName: sinkName, ChannelVolumes: info.ChannelVolumes}, nil); err != nil {
		return err
	}

	return nil
}

// SinkMuteToggle toggles a sink's muted status.
func (c *Client) SinkMuteToggle(sinkName string) error {
	info := proto.GetSinkInfoReply{}
	if err := c.proto.Request(&proto.GetSinkInfo{SinkIndex: proto.Undefined, SinkName: sinkName}, &info); err != nil {
		return err
	}

	if err := c.proto.Request(&proto.SetSinkMute{SinkIndex: proto.Undefined, SinkName: sinkName, Mute: !info.Mute}, nil); err != nil {
		return err
	}

	return nil
}

// SourceVolumeAdjust increases or decreases a source's volume.
func (c *Client) SourceVolumeAdjust(sourceName string, direction eventv1.Direction) error {
	info := proto.GetSourceInfoReply{}
	if err := c.proto.Request(&proto.GetSourceInfo{SourceIndex: proto.Undefined, SourceName: sourceName}, &info); err != nil {
		return err
	}

	for i, v := range info.ChannelVolumes {
		vol := int32(v)
		if direction == eventv1.Direction_DIRECTION_UP {
			if c.cfg.VolumeExceedMaximum && vol >= volumeMax {
				return nil
			} else if !c.cfg.VolumeExceedMaximum && vol >= int32(proto.VolumeNorm) {
				return nil
			}

			vol += int32(proto.VolumeNorm) / 100 * int32(c.cfg.VolumeStepPercent)
			if c.cfg.VolumeExceedMaximum && vol > volumeMax {
				vol = volumeMax
			} else if !c.cfg.VolumeExceedMaximum && vol > int32(proto.VolumeNorm) {
				vol = int32(proto.VolumeNorm)
			}
		} else {
			if vol <= 0 {
				return nil
			}

			vol -= int32(proto.VolumeNorm) / 100 * int32(c.cfg.VolumeStepPercent)
			if vol < 0 {
				vol = 0
			}
		}

		info.ChannelVolumes[i] = uint32(vol)
	}

	if info.Mute && direction == eventv1.Direction_DIRECTION_UP {
		if err := c.proto.Request(&proto.SetSourceMute{SourceIndex: proto.Undefined, SourceName: sourceName, Mute: false}, nil); err != nil {
			return err
		}
	}

	if err := c.proto.Request(&proto.SetSourceVolume{SourceIndex: proto.Undefined, SourceName: sourceName, ChannelVolumes: info.ChannelVolumes}, nil); err != nil {
		return err
	}

	return nil
}

// SourceMuteToggle toggles a source's muted status.
func (c *Client) SourceMuteToggle(sourceName string) error {
	info := proto.GetSourceInfoReply{}
	if err := c.proto.Request(&proto.GetSourceInfo{SourceIndex: proto.Undefined, SourceName: sourceName}, &info); err != nil {
		return err
	}

	if err := c.proto.Request(&proto.SetSourceMute{SourceIndex: proto.Undefined, SourceName: sourceName, Mute: !info.Mute}, nil); err != nil {
		return err
	}

	return nil
}

// Close the client.
func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) pollSink(idx uint32, name string) error {
	info := proto.GetSinkInfoReply{}
	if err := c.proto.Request(&proto.GetSinkInfo{SinkIndex: idx, SinkName: name}, &info); err != nil {
		return fmt.Errorf(`failed retrieving Pulse sink index (%d): %w`, idx, err)
	}
	var vol uint64
	for _, v := range info.ChannelVolumes {
		vol += uint64(v)
	}
	vol /= uint64(len(info.ChannelVolumes))

	sinkValue := &eventv1.AudioSinkChangeValue{
		Id:         info.SinkName,
		Name:       info.Device,
		Volume:     int32(vol),
		Percent:    math.Round(float64(vol)/float64(proto.VolumeNorm)*100) / 100,
		PercentMax: exceedVolumeFactor,
		Mute:       info.Mute,
	}

	if name == eventv1.AudioDefaultSink {
		sinkValue.Default = true
	} else {
		if err := c.proto.Request(&proto.GetSinkInfo{SinkIndex: proto.Undefined, SinkName: eventv1.AudioDefaultSink}, &info); err != nil {
			return fmt.Errorf(`failed retrieving default Pulse sink index (%d): %w`, idx, err)
		}
		if info.SinkIndex == idx || info.SinkName == name {
			sinkValue.Default = true
		}
	}

	if v, ok := c.cacheSink[info.SinkName]; ok {
		if v.Volume == sinkValue.Volume && v.Mute == sinkValue.Mute && v.Default == sinkValue.Default {
			return nil
		}
	}

	c.cacheSink[info.SinkName] = sinkValue

	sinkData, err := anypb.New(sinkValue)
	if err != nil {
		return fmt.Errorf(`failed encodiung Pulse event data for sink index (%d): %w`, idx, err)
	}

	c.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SINK_CHANGE,
		Data: sinkData,
	}

	if !c.cfg.HudNotifications {
		return nil
	}

	select {
	case <-c.readyCh:
	default:
		return nil
	}

	percent := math.Round(float64(vol)/float64(int32(proto.VolumeNorm))*100) / 100
	icon := `audio-volume-muted`
	if !info.Mute {
		switch {
		case percent >= 1:
			icon = `audio-volume-high`
		case percent >= 0.5:
			icon = `audio-volume-medium`
		case vol > 0:
			icon = `audio-volume-low`
		default:
			icon = `audio-volume-muted`
		}
	}

	hudValue := &eventv1.HudNotificationValue{
		Id:           hudID,
		Icon:         icon,
		IconSymbolic: true,
		Title:        info.Device,
		Body:         info.ActivePortName,
		Percent:      percent,
		PercentMax:   exceedVolumeFactor,
	}

	hudData, err := anypb.New(hudValue)
	if err != nil {
		return err
	}

	c.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_HUD_NOTIFY,
		Data: hudData,
	}

	return nil
}

func (c *Client) pollSource(idx uint32, name string) error {
	info := proto.GetSourceInfoReply{}
	if err := c.proto.Request(&proto.GetSourceInfo{SourceIndex: idx, SourceName: name}, &info); err != nil {
		return fmt.Errorf(`failed retrieving Pulse source index (%d): %w`, idx, err)
	}
	var vol uint64
	for _, v := range info.ChannelVolumes {
		vol += uint64(v)
	}
	vol /= uint64(len(info.ChannelVolumes))

	sourceValue := &eventv1.AudioSourceChangeValue{
		Id:         info.SourceName,
		Name:       info.Device,
		Volume:     int32(vol),
		Percent:    math.Round(float64(vol)/float64(proto.VolumeNorm)*100) / 100,
		PercentMax: exceedVolumeFactor,
		Mute:       info.Mute,
	}

	if name == eventv1.AudioDefaultSource {
		sourceValue.Default = true
	} else {
		if err := c.proto.Request(&proto.GetSourceInfo{SourceIndex: proto.Undefined, SourceName: eventv1.AudioDefaultSource}, &info); err != nil {
			return fmt.Errorf(`failed retrieving default Pulse source index (%d): %w`, idx, err)
		}
		if info.SourceIndex == idx || info.SourceName == name {
			sourceValue.Default = true
		}
	}

	if v, ok := c.cacheSource[info.SourceName]; ok {
		if v.Volume == sourceValue.Volume && v.Mute == sourceValue.Mute && v.Default == sourceValue.Default {
			return nil
		}
	}

	c.cacheSource[info.SourceName] = sourceValue

	sourceData, err := anypb.New(sourceValue)
	if err != nil {
		return fmt.Errorf(`failed encodiung Pulse event data for source index (%d): %w`, idx, err)
	}

	c.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_AUDIO_SOURCE_CHANGE,
		Data: sourceData,
	}

	if !c.cfg.HudNotifications {
		return nil
	}

	select {
	case <-c.readyCh:
	default:
		return nil
	}

	fraction := math.Round(float64(vol)/float64(int32(proto.VolumeNorm))*100) / 100
	icon := `audio-input-microphone-muted`
	if !info.Mute {
		switch {
		case fraction >= 1:
			icon = `audio-input-microphone-high`
		case fraction >= 0.5:
			icon = `audio-input-microphone-medium`
		case vol > 0:
			icon = `audio-input-microphone-low`
		default:
			icon = `audio-input-microphone-muted`
		}
	}

	hudValue := &eventv1.HudNotificationValue{
		Id:           hudID,
		Icon:         icon,
		IconSymbolic: true,
		Title:        info.Device,
		Body:         info.ActivePortName,
		Percent:      fraction,
		PercentMax:   exceedVolumeFactor,
	}

	hudData, err := anypb.New(hudValue)
	if err != nil {
		return err
	}

	c.eventCh <- &eventv1.Event{
		Kind: eventv1.EventKind_EVENT_KIND_HUD_NOTIFY,
		Data: hudData,
	}

	return nil
}

func (c *Client) init() error {
	c.proto.Callback = func(val any) {
		switch evt := val.(type) {
		case *proto.SubscribeEvent:
			c.log.Trace(`Pulse subscribe event received`, `index`, evt.Index, `type`, evt.Event.GetType(), `facility`, evt.Event.GetFacility())
			switch {
			case evt.Event.GetType() == proto.EventNew && evt.Event.GetFacility() == proto.EventSink:
			case evt.Event.GetType() == proto.EventChange && evt.Event.GetFacility() == proto.EventSink:
				select {
				case <-c.quitCh:
					return
				case c.sinkCh <- evt.Index:
				}
			case evt.Event.GetType() == proto.EventRemove && evt.Event.GetFacility() == proto.EventSink:
			case evt.Event.GetType() == proto.EventNew && evt.Event.GetFacility() == proto.EventSource:
			case evt.Event.GetType() == proto.EventChange && evt.Event.GetFacility() == proto.EventSource:
				select {
				case <-c.quitCh:
					return
				case c.sourceCh <- evt.Index:
				}
			case evt.Event.GetType() == proto.EventRemove && evt.Event.GetFacility() == proto.EventSource:
			case evt.Event.GetType() == proto.EventNew && evt.Event.GetFacility() == proto.EventCard:
			case evt.Event.GetType() == proto.EventChange && evt.Event.GetFacility() == proto.EventCard:
			case evt.Event.GetType() == proto.EventRemove && evt.Event.GetFacility() == proto.EventCard:
			}
		default:
			c.log.Trace(`Pulse unknown event received`, `evt`, evt)
		}
	}

	props := proto.PropList{
		`media.name`:                 proto.PropListString(pulseClientName),
		`application.name`:           proto.PropListString(pulseClientName),
		`application.icon_name`:      proto.PropListString(pulseIconName),
		`application.process.id`:     proto.PropListString(fmt.Sprintf("%d", os.Getpid())),
		`application.process.binary`: proto.PropListString(os.Args[0]),
	}

	if err := c.proto.Request(&proto.SetClientName{Props: props}, nil); err != nil {
		return fmt.Errorf("failed setting client properties: %w", err)
	}
	if err := c.proto.Request(&proto.Subscribe{Mask: proto.SubscriptionMaskAll}, nil); err != nil {
		return fmt.Errorf("failed enabling subscription: %w", err)
	}

	if err := c.pollSink(proto.Undefined, eventv1.AudioDefaultSink); err != nil {
		c.log.Error(`Failed retrieving default Pulse sink`, `err`, err)
	}

	if err := c.pollSource(proto.Undefined, eventv1.AudioDefaultSource); err != nil {
		c.log.Error(`Failed retrieving default Pulse source`, `err`, err)
	}

	close(c.readyCh)

	go c.watch()

	return nil
}

func (c *Client) watch() {
	for {
		select {
		case <-c.quitCh:
			return
		default:
			select {
			case <-c.quitCh:
				return
			case idx := <-c.sinkCh:
				if err := c.pollSink(idx, ``); err != nil {
					c.log.Error(`Failed processing Pulse event`, `err`, err)
				}
			case idx := <-c.sourceCh:
				if err := c.pollSource(idx, ``); err != nil {
					c.log.Error(`Failed processing Pulse event`, `err`, err)
				}
			}
		}
	}
}

// New instantiates a new pulseaudio client.
func New(cfg *configv1.Config_Audio, logger hclog.Logger) (*Client, <-chan *eventv1.Event, error) {
	p, conn, err := proto.Connect("")
	if err != nil {
		return nil, nil, fmt.Errorf("pulseaudio unavailable: %s", err)
	}

	c := &Client{
		cfg:         cfg,
		log:         logger,
		proto:       p,
		conn:        conn,
		cacheSink:   make(map[string]*eventv1.AudioSinkChangeValue),
		cacheSource: make(map[string]*eventv1.AudioSourceChangeValue),
		eventCh:     make(chan *eventv1.Event, 10),
		sinkCh:      make(chan uint32, 10),
		sourceCh:    make(chan uint32, 10),
		readyCh:     make(chan struct{}),
		quitCh:      make(chan struct{}),
	}

	if err := c.init(); err != nil {
		c.conn.Close()
		return nil, nil, err
	}

	return c, c.eventCh, nil
}
