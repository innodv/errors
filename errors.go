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
	"runtime"

	"github.com/jinzhu/copier"
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

func _copy(e interface{}) *Err {

	switch err := e.(type) {
	case *Err:
		if err == nil {
			return nil
		}
		return _copy(*err)
	case Err:
		out := new(Err)
		copier.Copy(&out, err)

		if err.Meta != nil {
			out.Meta = map[string]interface{}{}
			for k, v := range err.Meta {
				out.Meta[k] = v
			}
		}
		if err.Stack != nil {
			out.Stack = make([]Frame, len(err.Stack))
			for i := range out.Stack {
				copier.Copy(&out.Stack[i], err.Stack[i])
			}
		}
		return out
	}

	return nil
}

func _new(e interface{}, depth int) Error {
	if e == nil {
		return nil
	}
	var msg string
	switch val := e.(type) {
	case *Err:
		return _copy(*val)
	case error:
		msg = val.Error()
	default:
		msg = fmt.Sprintf("%v", e)
	}
	out := &Err{
		Meta:    map[string]interface{}{},
		Message: msg,
	}
	if depth != -1 {
		out.Stack = getStack(depth)
	}
	return out
}

func Plain(e interface{}) Error {
	return _new(e, -1)
}

func New(e interface{}) Error {
	return _new(e, 2)
}

func _wrap(err error, message string) Error {
	if err == nil {
		return nil
	}
	out := &Err{
		Err:     err,
		Message: message,
	}
	if e, ok := err.(*Err); ok {
		out.Stack = e.Stack
		out.Meta = e.Meta
	} else {
		out.Stack = getStack(2)
	}
	return out
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

func getStack(skip int) []Frame {
	pcs := make([]uintptr, StackBufferSize)
	length := runtime.Callers(2+skip, pcs)
	if length == 0 {
		return nil
	}
	pcs = pcs[:length]

	frames := runtime.CallersFrames(pcs)
	out := make([]Frame, 0, length)
	for {
		frame, more := frames.Next()

		if !more {
			break
		}
		fn := frame.Func.Name()
		if fn == "runtime.main" || fn == "runtime.goexit" {
			continue
		}
		out = append(out, Frame{
			Function: fn,
			File:     frame.File,
			Line:     frame.Line,
		})
	}
	return out
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
