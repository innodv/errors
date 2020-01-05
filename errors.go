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
	"sync"
	"sync/atomic"
	"unsafe"
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
	Err     error   `json:"error"`
	Message string  `json:"message"`
	Stack   []Frame `json:"stack"`
	meta    *sync.Map
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
		Message: msg,
		meta:    &sync.Map{},
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
		out.WithMeta(e.Meta())
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

func (err *Err) WithMeta(meta map[string]interface{}) Error {
	if err.meta == nil {
		mptr := &sync.Map{}
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&err.meta)), unsafe.Pointer(mptr))
	}
	for k, v := range meta {
		err.meta.Store(k, v)
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
	return errors.Is(err.Err, err2)
}

func (err *Err) Meta() map[string]interface{} {
	out := map[string]interface{}{}
	err.meta.Range(func(k, v interface{}) bool {
		out[k.(string)] = v
		return true
	})
	return out
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
	meta := err.Meta()
	if len(meta) > 0 {
		toJson["meta"] = meta
	}
	return json.Marshal(toJson)
}

func (err *Err) Unwrap() error {
	return err.Err
}
