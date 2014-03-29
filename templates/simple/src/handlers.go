// Copyright 2014 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
)

func (s *httpServer) route() {
	// Static file server.
	http.Handle("/static/", http.FileServer(http.Dir(s.config.DocumentRoot)))

	// Other handlers.
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/test", s.testHandler)
}

func (s *httpServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, world\r\n")
}

func (s *httpServer) testHandler(w http.ResponseWriter, r *http.Request) {
	// Set the "hello" key in redis first: redis-cli set hello world
	// Then call this handler: curl localhost:8080/test

	// The redis connection is fault-tolerant. Try killing redis and
	// calling /test again. Then run redis and call /test again.

	if v, err := s.redis.Get("hello"); err != nil {
		httpError(w, r, 503, err)
		return
	} else {
		fmt.Fprintf(w, "hello %s\r\n", v)
	}
}
