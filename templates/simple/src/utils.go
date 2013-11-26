// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

// remoteIP returns the remote IP without the port number.
func remoteIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		return r.RemoteAddr // xheaders?
	} else {
		return ip
	}
	return "" // Go1.0
}

// serverURL returns the base URL of the server based on the current request.
func serverURL(r *http.Request, preferSSL bool) string {
	var (
		addr  string
		host  string
		port  string
		proto string
	)
	if cfg.HTTPS.Addr == "" || !preferSSL {
		proto = "http"
		addr = cfg.HTTP.Addr
	} else {
		proto = "https"
		addr = cfg.HTTPS.Addr
	}
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			port = addr[i+1:]
			break
		}
	}
	host = r.Host
	if port != "" {
		for i := len(host) - 1; i >= 0; i-- {
			if host[i] == ':' {
				host = host[:i]
				break
			}
		}
		if port != "80" && port != "443" {
			host += ":" + port
		}
	}
	return fmt.Sprintf("%s://%s/", proto, host)
}

// httpError renders the default error message based on
// the status code, and prints the program error to the log.
func httpError(w http.ResponseWriter, code int, msg ...interface{}) {
	http.Error(w, http.StatusText(code), code)
	if msg != nil && len(msg) >= 1 {
		switch msg[0].(type) {
		case string:
			log.Printf(msg[0].(string), msg[1:]...)
		case nil:
			// ignore
		default:
			log.Println("Error", msg)
		}
	}
}

// NewJSON encodes `d` as JSON and writes it to the http connection.
func NewJSON(w http.ResponseWriter, d interface{}) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

// JSON reads the HTTP request body and parses it as JSON.
func JSON(r *http.Request, v interface{}) error {
	// TODO: check mime type first?
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &v)
}
