package eventv1

import (
	"errors"
	"strconv"

	anypb "google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var ErrTypeMismatch = errors.New(`requested type did not match data type`)

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
