// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

// unixsocket.go starts the HTTP server on a Unix socket instead of TCP.
// Useful when the server is reverse proxied by Nginx or other servers.
// Make sure the frontend server sets the X-Forwarded-For HTTP header with the
// IP address of the client.
// http.Server must be created manually, with the XHeaders option set to true.
// When XHeaders is set to true, it overwrites http.Request.RemoteAddr with
// the contents of X-Forwarded-For HTTP header.
// It does not validate the IP.
// Test:
// echo -ne 'GET / HTTP/1.1\r\nX-Forwarded-For: pwnz\r\n\r\n' | nc -U ./test.sock

import (
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/fiorix/go-web/http"
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
	http.HandleFunc("/", IndexHandler)

	// Try to delete the socket first, otherwise ListenAndServe fails
	// and returns an error like "address already in use".
	syscall.Unlink("./test.sock")

	// Create and start the server
	server := http.Server{
		Addr:   "unix:./test.sock", // Listen on Unix Socket
		Logger: logger,             // Logger to be called after every request

		// XHeaders make the server overwrite the remote IP address in
		// http.Request.RemoteAddr with the contents of the X-Forwarded-For
		// HTTP header when possible.
		XHeaders: true,
	}
	if e := server.ListenAndServe(); e != nil {
		fmt.Println(e.Error())
	}
}
