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
