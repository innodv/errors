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

	errs "github.com/pkg/errors"
)

type Frame struct {
	Function string `json:"function,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

const StackBufferSize = 100

type Error interface {
	Error() string
	WithMeta(meta map[string]interface{}) Error
	WithStack() Error
	Is(err error) bool
	MarshalJSON() ([]byte, error)
}

type Err struct {
	Err   error                  `json:"error"`
	Stack []Frame                `json:"stack"`
	Meta  map[string]interface{} `json:"meta"`
}

func _new(e interface{}, depth int) Error {
	out := &Err{
		Stack: getStack(depth),
		Meta:  map[string]interface{}{},
	}
	if err, ok := e.(error); ok {
		out.Err = err
	} else {
		out.Err = fmt.Errorf("%v", e)
	}

	return out
}

func New(e interface{}) Error {
	return _new(e, 2)
}

func Wrap(err error, message string) Error {
	if e, ok := err.(*Err); ok {
		e.Err = errs.Wrap(err, message)
		return e
	}
	return _new(errs.Wrap(err, message), 2)
}

func (err *Err) Error() string {
	return err.Err.Error()
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
	if e, ok := err2.(*Err); ok {
		return errors.Is(err.Err, e.Err)
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
	toJson := map[string]interface{}{
		"error": err.Err.Error(),
		"stack": err.Stack,
		"meta":  err.Meta,
	}
	return json.Marshal(toJson)
}
