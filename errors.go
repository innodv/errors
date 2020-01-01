/**
 * Copyright 2019 Innodev LLC. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package errors

import (
	"encoding/json"
	errs "errors"
	"fmt"
	"runtime"
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

func _new(e interface{}, depth int) Error {
	if e == nil {
		return nil
	}
	var msg string
	switch val := e.(type) {
	case *Err:
		return val
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

func (err *Err) Error() string {
	if err == nil {
		return "<nil>"
	}
	if err.Err == nil {
		return err.Message
	}
	return err.Err.Error() + ":" + err.Message
}

func (err *Err) WithMeta(meta map[string]interface{}) Error {
	for k, v := range meta {
		err.Meta[k] = v
	}

	return err
}

func (err *Err) WithStack() Error {
	err.Stack = getStack(1)
	return err
}

func (err *Err) Is(err2 error) bool {
	if err.Message == err2.Error() {
		return true
	}
	return errs.Is(err.Err, err2)
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
	toJson := map[string]interface{}{
		"error": err.Error(),
		"stack": err.Stack,
		"meta":  err.Meta,
	}
	return json.Marshal(toJson)
}

func (err *Err) Unwrap() error {
	return err.Err
}
