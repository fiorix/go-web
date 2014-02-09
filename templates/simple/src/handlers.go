// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, world\r\n")
}

func TestHandler(w http.ResponseWriter, r *http.Request) {

	// Run this on the command line:
	// redis-cli set foo bar

	if bar, err := Redis.Get("foo"); err != nil {
		httpError(w, r, 503, "Redis: "+err.Error())
		return
	} else {
		fmt.Fprintf(w, "foo:%s\r\n", bar)
	}
}
