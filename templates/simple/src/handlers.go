// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	bar, _ := Redis.Get("foo") // redis-cli set foo bar
	fmt.Fprintln(w, "Hello, world", bar)
}
