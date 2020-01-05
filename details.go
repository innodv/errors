/**
 * Copyright 2019 Innodev LLC. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package errors

import (
	"fmt"
	"runtime"

	"github.com/jinzhu/copier"
)

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

func _wrap(err error, message string) Error {
	if err == nil {
		return nil
	}
	out := &Err{
		Message: message,
	}
	if e, ok := err.(*Err); ok {
		cpy := _copy(e)
		out.Err = cpy
		out.Stack = cpy.Stack
		out.Meta = cpy.Meta
	} else {
		out.Stack = getStack(2)
	}
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
