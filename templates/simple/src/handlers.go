// Copyright 2013 %template% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world")

	// Run this on the command line: redis-cli set foo bar
	if bar, err := Redis.Get("foo"); err != nil {
		fmt.Fprintln(w, "foo:", bar)
	}
}
