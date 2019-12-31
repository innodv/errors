package errors

import (
	"errors"
	"fmt"
	"runtime"
)

type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

const StackBufferSize = 100

type Error interface {
	Error() string
	WithMeta(meta map[string]interface{}) Error
	WithStack() Error
	Is(err error) bool
}

type Err struct {
	Err   error                  `json:"error"`
	Stack []Frame                `json:"stack"`
	Meta  map[string]interface{} `json:"meta"`
}

func New(e interface{}) Error {
	out := &Err{
		Stack: getStack(1),
		Meta:  map[string]interface{}{},
	}
	if err, ok := e.(error); ok {
		out.Err = err
	} else {
		out.Err = fmt.Errorf("%v", e)
	}

	return out
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
		out = append(out, Frame{
			Function: frame.Func.Name(),
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
	}
	return out
}
