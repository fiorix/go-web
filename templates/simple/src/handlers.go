// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {

	// Run this on the command line:
	// redis-cli set foo bar

	if bar, err := Redis.Get("foo"); err != nil {
		httpError(w, 503, "Redis: "+err.Error())
		return
	} else {
		fmt.Fprintln(w, "foo:", bar)
	}

}
