/**
 * Copyright 2019 Innodev LLC. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package errors

import (
	errs "errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	err := New("foobar")
	stack := err.(*Err).Stack
	err = Wrap(err, io.EOF.Error())
	err = Wrap(err, "water buffalo")
	assert.ElementsMatch(t, stack, err.(*Err).Stack)
	assert.True(t, errs.Is(err, io.EOF))
	assert.True(t, Is(err, io.EOF))
	assert.Equal(t, err.Error(), "foobar:EOF:water buffalo")

}

func TestErrorsImmutable(t *testing.T) {
	err := New("foobar").WithMeta(map[string]interface{}{
		"hello": "world",
	})

	err2 := err.WithMeta(map[string]interface{}{
		"world": "hello",
	})

	err3 := err2.WithStack()

	err4 := Wrap(err3, "test")

	assert.Equal(t, "world", err2.(*Err).Meta["hello"])
	assert.Equal(t, "world", err.(*Err).Meta["hello"])

	_, ok := err.(*Err).Meta["world"]
	assert.False(t, ok)
	assert.NotEqual(t, err, err2)

	assert.NotEqual(t, err2, err3)

	assert.Equal(t, err4.Unwrap().Error(), err3.Error())
	err4.(*Err).Err.(*Err).Meta["foo"] = "bar"
	assert.NotEqual(t, err4.(*Err).Err, err3)
}
