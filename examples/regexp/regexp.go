// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

// Custom multiplexer and logger

import (
	"fmt"
	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/remux"
	"log"
	"time"
)

func logger(w http.ResponseWriter, req *http.Request) {
	log.Printf("HTTP %d %s %s (%s) :: %s",
		w.Status(),
		req.Method,
		req.URL.Path,
		req.RemoteAddr,
		time.Since(req.Created))
}

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello, world")
}

func main() {
	// Supports GET /, /foo and /bar
	remux.HandleFunc("^/(foo|bar)?$", IndexHandler)

	// Create and start the server
	server := http.Server{
		Addr:    ":8080",
		Handler: remux.DefaultServeMux,
		Logger:  logger,
	}
	server.ListenAndServe()
}
