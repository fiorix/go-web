// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/fiorix/go-web/httpxtra"
)

func RouteHTTP() {
	// Public handlers: add your own
	http.Handle("/", http.FileServer(http.Dir(cfg.DocumentRoot)))
	http.HandleFunc("/test", TestHandler)
}

func ListenHTTP() {
	s := http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: httpxtra.Handler{Logger: _logger},
	}
	log.Println("Starting HTTP server on", cfg.HTTP.Addr)
	log.Fatal(s.ListenAndServe())
}

func ListenHTTPS() {
	s := http.Server{
		Addr:    cfg.HTTPS.Addr,
		Handler: httpxtra.Handler{Logger: _logger},
	}
	log.Println("Starting HTTPS server on", cfg.HTTPS.Addr)
	log.Fatal(s.ListenAndServeTLS(cfg.HTTPS.CertFile, cfg.HTTPS.KeyFile))
}

func _logger(r *http.Request, created time.Time, status, bytes int) {
	//fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))
	if r.TLS == nil {
		log.Printf("HTTP %d %s %q (%s) :: %s",
			status,
			r.Method,
			r.URL.Path,
			remoteIP(r),
			time.Since(created),
		)
	} else {
		log.Printf("HTTPS %d %s %q (%s) :: %s",
			status,
			r.Method,
			r.URL.Path,
			remoteIP(r),
			time.Since(created),
		)
	}
}
