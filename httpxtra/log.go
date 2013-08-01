// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package httpxtra

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)

// LoggerFunc are functions called by httpxtra.Handler at the end of each request.
type LoggerFunc func(r *http.Request, created time.Time, status, bytes int)

type logWriter struct {
	w      http.ResponseWriter
	bytes  int
	status int
}

func (lw *logWriter) Header() http.Header {
	return lw.w.Header()
}

func (lw *logWriter) Write(b []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	n, err := lw.w.Write(b)
	lw.bytes += n
	return n, err
}

func (lw *logWriter) WriteHeader(s int) {
	lw.w.WriteHeader(s)
	lw.status = s
}

func (lw *logWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	// TODO: Check. Does it break if the server don't support hijacking?
	return lw.w.(http.Hijacker).Hijack()
}

// ApacheCommonLog returns an Apache Common access log string.
func ApacheCommonLog(r *http.Request, created time.Time, status, bytes int) string {
	u := "-"
	if r.URL.User != nil {
		if name := r.URL.User.Username(); name != "" {
			u = name
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d",
		ip,
		u,
		created.Format("02/Jan/2006:15:04:05 -0700"),
		r.Method,
		r.RequestURI,
		r.Proto,
		status,
		bytes,
	)
}
