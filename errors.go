/**
 * Copyright 2019 Innodev LLC. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package errors

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Frame struct {
	Function string `json:"function,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

func (frame Frame) String() string {
	return fmt.Sprintf("\"%s:%d %s()\"", frame.File, frame.Line, frame.Function)
}

const StackBufferSize = 100

type Error interface {
	Error() string
	WithMeta(meta map[string]interface{}) Error
	WithStack() Error
	Is(err error) bool
	MarshalJSON() ([]byte, error)
	Unwrap() error
}

type Err struct {
	Err     error                  `json:"error"`
	Message string                 `json:"message"`
	Stack   []Frame                `json:"stack"`
	Meta    map[string]interface{} `json:"meta"`
}

func Plain(e interface{}) Error {
	return _new(e, -1)
}

func New(e interface{}) Error {
	return _new(e, 2)
}

func Wrap(err error, message string) Error {
	return _wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return _wrap(err, fmt.Sprintf(format, args...))
}

// Standard library passthroughs to allow for use as drop-in replacement

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func (err *Err) Error() string {
	if err == nil {
		return "<nil>"
	}
	if err.Err == nil {
		return err.Message
	}
	return err.Err.Error() + ":" + err.Message
}

func (err Err) WithMeta(meta map[string]interface{}) Error {
	out := _copy(err)
	if out.Meta == nil {
		out.Meta = map[string]interface{}{}
	}
	for k, v := range meta {
		out.Meta[k] = v
	}

	return out
}

func (err Err) WithStack() Error {
	out := _copy(err)
	out.Stack = getStack(1)
	return out
}

func (err Err) Is(err2 error) bool {
	if err.Message == err2.Error() {
		return true
	}
	return errors.Is(err.Err, err2)
}

func (err *Err) MarshalJSON() ([]byte, error) {
	if err == nil {
		return json.Marshal(nil)
	}
	toJson := map[string]interface{}{
		"error": err.Error(),
	}
	if len(err.Stack) > 0 {
		toJson["stack"] = err.Stack
	}
	if len(err.Meta) > 0 {
		toJson["meta"] = err.Meta
	}
	return json.Marshal(toJson)
}

func (err *Err) Unwrap() error {
	return err.Err
}

func MustNotError(err error) {
	if err != nil {
		panic(err)
	}
}
