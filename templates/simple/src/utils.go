// Copyright 2014 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// remoteIP returns the remote IP without the port number.
func remoteIP(r *http.Request) string {
	// If xheaders is enabled, RemoteAddr might be a copy of
	// the X-Real-IP or X-Forwarded-For HTTP headers, which
	// can be a comma separated list of IPs. In this case,
	// only the first IP in the list is used.
	if strings.Index(r.RemoteAddr, ",") > 0 {
		r.RemoteAddr = strings.SplitN(r.RemoteAddr, ",", 2)[0]
	}
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
		return ip
	} else {
		return r.RemoteAddr
	}
}

// serverURL returns the base URL of the server based on the current request.
func serverURL(config *configFile, r *http.Request, preferSSL bool) string {
	var (
		addr  string
		host  string
		port  string
		proto string
	)
	if config.HTTPS.Addr == "" || !preferSSL {
		proto = "http"
		addr = config.HTTP.Addr
	} else {
		proto = "https"
		addr = config.HTTPS.Addr
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
