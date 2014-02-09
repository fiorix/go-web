// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/fiorix/go-web/httpxtra"
	"github.com/gorilla/context"
)

func RouteHTTP() {
	// Static file server.
	http.Handle("/static/", http.FileServer(http.Dir(Config.DocumentRoot)))

	// Public handlers: add your own
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/test", TestHandler)
}

func ListenHTTP() {
	s := http.Server{
		Addr:    Config.HTTP.Addr,
		Handler: httpxtra.Handler{Logger: httpLogger},
	}
	log.Println("Starting HTTP server on", Config.HTTP.Addr)
	log.Fatal(s.ListenAndServe())
}

func ListenHTTPS() {
	s := http.Server{
		Addr:    Config.HTTPS.Addr,
		Handler: httpxtra.Handler{Logger: httpLogger},
	}
	log.Println("Starting HTTPS server on", Config.HTTPS.Addr)
	log.Fatal(s.ListenAndServeTLS(Config.HTTPS.CertFile, Config.HTTPS.KeyFile))
}

// httpError renders the default error message based on
// the status code, and sets the "info" context variable with the error.
func httpError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	// TODO: render error page instead of text?
	http.Error(w, http.StatusText(code), code)

	if msg != "" {
		context.Set(r, "info", msg)
	}
}

// httpLogger is called at the end of every HTTP request. It dumps one
// log line per request.
//
// The "info" context variable can be used to add extra information to
// the logging, such as database or template errors.
func httpLogger(r *http.Request, created time.Time, status, bytes int) {
	//fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))

	var proto, info string

	if r.TLS == nil {
		proto = "HTTP"
	} else {
		proto = "HTTPS"
	}

	if tmp := context.Get(r, "info"); tmp != nil {
		info = "(" + tmp.(string) + ")"
	}

	log.Printf("%s %d %s %q (%s) :: %s %s",
		proto,
		status,
		r.Method,
		r.URL.Path,
		remoteIP(r),
		time.Since(created),
		info,
	)
}
