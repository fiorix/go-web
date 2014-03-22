// Copyright 2013-2014 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// remoteIP returns the remote IP without the port number.
func remoteIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		// If xheaders is enabled, RemoteAddr might be a copy of
		// the X-Real-IP or X-Forwarded-For HTTP headers, which
		// can be a comma separated list of IPs. In this case,
		// only the first IP in the list is used.
		return strings.SplitN(r.RemoteAddr, ",", 2)[0]
	} else {
		return ip
	}
}

// serverURL returns the base URL of the server based on the current request.
func serverURL(r *http.Request, preferSSL bool) string {
	var (
		addr  string
		host  string
		port  string
		proto string
	)
	if Config.HTTPS.Addr == "" || !preferSSL {
		proto = "http"
		addr = Config.HTTP.Addr
	} else {
		proto = "https"
		addr = Config.HTTPS.Addr
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
	return fmt.Sprintf("%s://%s", proto, host)
}

// writeJSON encodes `d` as JSON and writes it to the http connection.
func writeJSON(w http.ResponseWriter, d interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(d)
}

// readJSON reads the HTTP request body and parses it as JSON.
func readJSON(r *http.Request, v interface{}) error {
	// TODO: check mime type first?
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}
