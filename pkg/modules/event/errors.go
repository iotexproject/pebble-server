package event

import (
	"encoding/hex"
	"fmt"
	"reflect"
)

func WrapUnmarshalError(e error, t any) error {
	if e == nil {
		return nil
	}
	return &UnmarshalError{t, e}
}

type UnmarshalError struct {
	t any
	e any
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf(
		"failed to unmarshal payload for `%s`: [err:%v]",
		reflect.Indirect(reflect.ValueOf(e.t)).Type(), e.e,
	)
}

type UnmarshalTopicError struct {
	topic string
	event any
}

func (e *UnmarshalTopicError) Error() string {
	return fmt.Sprintf(
		"failed to unmarshal topic for `%s` from `%s`",
		reflect.Indirect(reflect.ValueOf(e.event)).Type(), e.topic,
	)
}

func WrapValidateError(v CanValidateSignature) *ValidateError {
	return &ValidateError{v}
}

type ValidateError struct {
	v CanValidateSignature
}

func (e *ValidateError) Error() string {
	return fmt.Sprintf(
		"failed to validate signature for `%s`: [hash: %s] [addr: %s] [sig: %s]",
		reflect.Indirect(reflect.ValueOf(e.v)).Type(),
		hex.EncodeToString(e.v.Hash()), e.v.Address(), hex.EncodeToString(e.v.Signature()),
	)
}

func WrapHandleError(e error, t any) error {
	if e == nil {
		return nil
	}
	return &HandleError{t, e, ""}
}

func WrapHandleErrorf(e error, t any, msg string, args ...any) error {
	if e == nil {
		return nil
	}
	if len(args) > 0 && msg != "" {
		msg = fmt.Sprintf(msg, args...)
	}
	return &HandleError{t, e, msg}
}

type HandleError struct {
	t   any
	e   error
	msg string
}

func (e *HandleError) Error() string {
	msg := fmt.Sprintf(
		"failed to handle event `%s`: %s",
		reflect.Indirect(reflect.ValueOf(e.t)).Type(), e.e.Error(),
	)
	if e.msg != "" {
		return msg + " " + e.msg
	}
	return msg
}
