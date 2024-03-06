package eventv1

import (
	"errors"
	"strconv"

	anypb "google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ErrTypeMismatch is returned when operating on incompatible protobuf message type.
var ErrTypeMismatch = errors.New(`requested type did not match data type`)

// NewString convenience function for instantiating an event with a string data attribute.
func NewString(kind EventKind, value string) (*Event, error) {
	data, err := anypb.New(wrapperspb.String(value))
	if err != nil {
		return nil, err
	}
	return &Event{
		Kind: kind,
		Data: data,
	}, nil
}

// NewInt32 convenience function for instantiating an event with an int32 data attribute.
func NewInt32(kind EventKind, value string) (*Event, error) {
	v, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	data, err := anypb.New(wrapperspb.Int32(int32(v)))
	if err != nil {
		return nil, err
	}
	return &Event{
		Kind: kind,
		Data: data,
	}, nil
}

// NewUInt32 convenience function for instantiating an event with a uint32 data attribute.
func NewUInt32(kind EventKind, value string) (*Event, error) {
	v, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return nil, err
	}
	data, err := anypb.New(wrapperspb.UInt32(uint32(v)))
	if err != nil {
		return nil, err
	}
	return &Event{
		Kind: kind,
		Data: data,
	}, nil
}

// DataString convenience function for extracting a value from a string anypb field.
func DataString(a *anypb.Any) (string, error) {
	v := &wrapperspb.StringValue{}
	if !a.MessageIs(v) {
		return ``, ErrTypeMismatch
	}
	if err := a.UnmarshalTo(v); err != nil {
		return ``, err
	}
	return v.Value, nil

}

// DataInt32 convenience function for extracting a value from an int32 anypb field.
func DataInt32(a *anypb.Any) (int32, error) {
	v := &wrapperspb.Int32Value{}
	if !a.MessageIs(v) {
		return 0, ErrTypeMismatch
	}
	if err := a.UnmarshalTo(v); err != nil {
		return 0, err
	}
	return v.Value, nil
}

// DataInt64 convenience function for extracting a value from an int64 anypb field.
func DataInt64(a *anypb.Any) (int64, error) {
	v := &wrapperspb.Int64Value{}
	if !a.MessageIs(v) {
		return 0, ErrTypeMismatch
	}
	if err := a.UnmarshalTo(v); err != nil {
		return 0, err
	}
	return v.Value, nil
}

// DataUInt32 convenience function for extracting a value from a uint32 anypb field.
func DataUInt32(a *anypb.Any) (uint32, error) {
	v := &wrapperspb.UInt32Value{}
	if !a.MessageIs(v) {
		return 0, ErrTypeMismatch
	}
	if err := a.UnmarshalTo(v); err != nil {
		return 0, err
	}
	return v.Value, nil
}
