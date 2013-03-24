// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Same as the default net/http.

package main

import (
	"fmt"
	"github.com/fiorix/go-web/http"
)

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello, world")
}

func main() {
	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe(":8080", nil)
}
