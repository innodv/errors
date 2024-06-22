/**
 * Copyright 2019 Innodev LLC. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package await

import (
	"github.com/pkg/errors"
)

func AwaitErrors(errChan <-chan error, count int) error {
	var err error
	for i := 0; i < count; i++ {
		err2 := <-errChan
		if err2 != nil {
			if err == nil {
				err = err2
			} else {
				err = errors.Wrap(err, err2.Error())
			}
		}
	}
	return err
}
