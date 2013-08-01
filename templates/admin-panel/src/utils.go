// Copyright 2013 %template% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"

	"github.com/gorilla/sessions"
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

// serverURL returns the URL of the server based on the current request.
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
	return fmt.Sprintf("%s://%s/", proto, host)
}

// nocsrf protects against csrf or xsrf attacks.
func nocsrf(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Requested-With") == "" {
			http.NotFound(w, r)
		} else {
			fn(w, r)
		}
	}
}

type AuthHandlerFunc func(http.ResponseWriter, *http.Request, *sessions.Session)

// authenticated is a wrapper for HandlerFunc functions that automatically
// checks the (cookie) session. If there's no session available then the
// request is automatically redirected to the sign in page.
func authenticated(fn AuthHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := Session.Get(r, "s")
		if s == nil || s.Values["Id"] == nil || err != nil {
			http.Redirect(w, r, "/signin.html", 302)
			return
		}
		fn(w, r, s)
	}
}

// UserFS is an http.FileServer for authenticated sessions only.
func UserFS(dir string) AuthHandlerFunc {
	fs := http.FileServer(http.Dir(Config.UsersDocumentRoot))
	return func(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
		fs.ServeHTTP(w, r)
	}
}

// httpError renders the default error message to the http client based on
// the code, and prints the program error to the log.
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

// RandHex generates random hex sequences.
// TODO: Someone fix for Windows.
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

// SendMail sends an email to the pre-configured SMTP server.
// TODO: Fix the time out.
// If the SMTP server is unreachable SendMail takes a long time to return.
// In the real world sendMail should just drop the msg into a queue.
func SendMail(to []string, msg []byte) error {
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

// JSON encodes a message as JSON and writes to the socket.
func JSON(w http.ResponseWriter, d interface{}) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = io.Copy(w, bytes.NewReader(b))
	return err
}

// ParseJSON reads an HTTP request body and parses its JSON content.
func ParseJSON(r *http.Request, v interface{}) error {
	// TODO: check mime type first?
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &v)
}
