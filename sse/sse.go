// Copyright 2013-2014 The go-web authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Server-Sent events (SSE)
// http://dev.w3.org/html5/eventsource/
//
// Usage example:
//
//	func SSEHandler(w http.ResponseWriter, req *http.Request) {
//	        conn, err := sse.ServeEvents(w)
//	        if err != nil {
//	                http.Error(w, err.Error(), http.StatusInternalServerError)
//	                return
//	        }
//	        defer conn.Close()
//	        for i := 0; i < 10; i++ {
//	                sse.SendEvent(conn, &sse.MessageEvent{Data: "Hello, world"})
//	                time.Sleep(1 * time.Second)
//	        }
//	}
package sse

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
)

var ErrNoHijack = errors.New("Server does not support hijacking")

// MessageEvent is the container of Server-Sent events (SSE), push notifications.
type MessageEvent struct {
	Data  string // message content
	Id    string // id of the message (int?)
	Event string // name of the event
	Retry int    // client reconnection time, in milliseconds
}

// ServeEvents prepares the request for SSE, push notifications.
// Caveat: ResponseWriter.Status() returns 0 instead of 200 after ServeEvents
// is called, and might break logging.
func ServeEvents(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, nil, ErrNoHijack
	}
	conn, rw, err := hj.Hijack()
	if err != nil {
		return nil, nil, err
	}
	fmt.Fprintf(rw, "HTTP/1.1 200 OK\r\n")
	w.Header().Write(rw)
	fmt.Fprintf(rw, "\r\n")
	return conn, rw, rw.Flush()
}

// SendEvent sends a push notification to the peer (usually a browser).
// Browsers can handle these events in JavaScript:
// http://www.w3schools.com/html/html5_serversentevents.asp
func SendEvent(rw *bufio.ReadWriter, m *MessageEvent) error {
	if m.Data != "" {
		fmt.Fprintf(rw, "data: %s\n", m.Data)
	}
	if m.Event != "" {
		fmt.Fprintf(rw, "event: %s\n", m.Event)
	}
	if m.Id != "" {
		fmt.Fprintf(rw, "id: %s\n", m.Id)
	}
	if m.Retry >= 1 {
		fmt.Fprintf(rw, "retry: %d\n", m.Retry)
	}
	fmt.Fprintf(rw, "\n")
	return rw.Flush()
}
