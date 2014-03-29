// Copyright 2014 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fiorix/go-web/httpxtra"
	"github.com/gorilla/context"

	"github.com/fiorix/go-redis/redis"
)

type httpServer struct {
	config *configFile
	redis  *redis.Client
	mysql  *sql.DB
}

func (s *httpServer) init(cf *configFile, rc *redis.Client, db *sql.DB) {
	s.config = cf
	s.redis = rc
	s.mysql = db

	// Initialize http handlers.
	s.route()
}

func (s *httpServer) ListenAndServe() {
	if s.config.HTTP.Addr == "" {
		return
	}
	srv := http.Server{
		Addr:    s.config.HTTP.Addr,
		Handler: httpxtra.Handler{Logger: httpLogger},
	}
	log.Println("Starting HTTP server on", s.config.HTTP.Addr)
	log.Fatal(srv.ListenAndServe())
}

func (s *httpServer) ListenAndServeTLS() {
	if s.config.HTTPS.Addr == "" {
		return
	}
	srv := http.Server{
		Addr:    s.config.HTTPS.Addr,
		Handler: httpxtra.Handler{Logger: httpLogger},
	}
	log.Println("Starting HTTPS server on", s.config.HTTPS.Addr)
	log.Fatal(srv.ListenAndServeTLS(
		s.config.HTTPS.CertFile,
		s.config.HTTPS.KeyFile,
	))
}

// httpError renders the default error message based on
// the status code, and sets the "log" context variable with the error.
func httpError(w http.ResponseWriter, r *http.Request, code int, msg interface{}) {
	http.Error(w, http.StatusText(code), code)
	if msg != nil {
		context.Set(r, "log", msg)
	}
}

// httpLogger is called at the end of every HTTP request. It dumps one
// log line per request.
//
// The "log" context variable can be set by handlers to add extra
// information to the log message, such as database or template errors.
func httpLogger(r *http.Request, created time.Time, status, bytes int) {
	//fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))

	log.Printf("%s %d %s %q (%s) :: %d bytes in %s%s",
		logProto(r),
		status,
		r.Method,
		r.URL.Path,
		remoteIP(r),
		bytes,
		time.Since(created),
		logMsg(r),
	)
}

func logProto(r *http.Request) string {
	if r.TLS == nil {
		return "HTTP"
	} else {
		return "HTTPS"
	}
}

func logMsg(r *http.Request) string {
	if msg := context.Get(r, "log"); msg != nil {
		defer context.Clear(r)
		return fmt.Sprintf(" (%s)", msg)
	}
	return ""
}
