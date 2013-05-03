// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// httpxtra is wrapper for http.Handler that adds extra features to the server:
// - Custom logging
// - Support for listening on TCP or UNIX sockets
// - Support X-Real-IP and X-Forwarded-For as the remote IP if the server sits
//   behind a proxy or load balancer.
package httpxtra

import (
	"net/http"
	"time"
)

// Handler is the http.Handler wrapper with extra features.
type Handler struct {
	Handler  http.Handler
	Logger   LoggerFunc
	XHeaders bool
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	lw := logWriter{w: w}
	if h.Handler == nil {
		h.Handler = http.DefaultServeMux
	}
	if h.XHeaders {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip != "" {
			r.RemoteAddr = ip
		}
	}
	h.Handler.ServeHTTP(&lw, r)
	if h.Logger != nil {
		h.Logger(r, t, lw.status, lw.bytes)
	}
}
