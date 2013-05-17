// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"

	"github.com/gorilla/sessions"
)

// https is a wrapper for HandlerFunc functions that enforces HTTPS on GET.
func https(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil && Config.SSL.Addr != "" && r.Method == "GET" {
			url, _, _ := net.SplitHostPort(r.Host)
			_, port, _ := net.SplitHostPort(Config.SSL.Addr)
			if port != "" && port != "443" {
				url += ":" + port
			}
			url += r.URL.Path
			if r.URL.RawQuery != "" {
				url += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, "https://"+url, 302)
		} else {
			fn(w, r)
		}
	}
}

type SessHandlerFunc func(http.ResponseWriter, *http.Request, *sessions.Session)

// authenticated is a wrapper for HandlerFunc functions that automatically
// checks the (cookie) session. If there's no session available then the
// request is automatically redirected to /signin?next=current_url.
func authenticated(fn SessHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil && Config.SSL.Addr != "" {
			http.Redirect(w, r, "https://", 302)
		}
		s, err := Session.Get(r, "s")
		if s == nil || s.Values["Id"] == nil || err != nil {
			http.Redirect(w, r, "/signin/?next="+r.URL.Path, 302)
			return
		}
		fn(w, r, s)
	}
}

// unauthenticated is a wrapper for HandlerFunc functions that prevents
// signed in users to access certain endpoints, such as sign in and sign up.
func unauthenticated(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s, _ := Session.Get(r, "s"); s != nil && s.Values["Id"] != nil {
			http.Redirect(w, r, "/main/", 302)
			return
		}
		fn(w, r)
	}
}

// httpError renders a custom error page and prints a message to the log.
// TODO: Render a custom error page.
func httpError(w http.ResponseWriter, code int, msg ...interface{}) {
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
	http.Error(w, http.StatusText(code), code)
}

// RandHex generates random hex sequences.
// TODO: Fix for Windows users.
func RandHex(nbytes int) string {
	file, err := os.Open("/dev/urandom")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	bytes := make([]byte, nbytes)
	_, err = file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(bytes)
}

// staticFile serves a static file under the StaticFile configuration entry.
func staticFile(name string) string {
	return filepath.Join(Config.StaticPath, name)
}

// sendMail sends an email to the pre-configured SMTP server.
// TODO: Fix the time out.
//       When the server is unreachable SendMail takes a long time to return.
func sendMail(to []string, msg []byte) error {
	return smtp.SendMail(
		Config.SMTP.Addr,
		smtp.PlainAuth(
			"", // Identity
			Config.SMTP.PlainAuth.User,
			Config.SMTP.PlainAuth.Passwd,
			Config.SMTP.PlainAuth.Host),
		Config.SMTP.From,
		to, msg)
}
