// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

// Custom multiplexer and logger

import (
	"fmt"
	"log"
	"time"

	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/remux"
)

func logger(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP %d %s %s (%s) :: %s",
		w.Status(),
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		time.Since(r.Created))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
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
