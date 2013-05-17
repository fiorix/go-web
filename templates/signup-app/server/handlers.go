// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func StaticHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path[:7] == "/static" {
		http.ServeFile(w, r, staticFile(r.URL.Path[7:]))
	} else {
		http.ServeFile(w, r, staticFile(r.URL.Path))
	}
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "signin.html", nil)
	case "POST":
		next := r.FormValue("next")
		email := r.FormValue("email")
		passwd := r.FormValue("passwd")
		remember := r.FormValue("remember")
		u, err := GetUserWithPasswd(email, passwd)
		if err != nil {
			if err == sql.ErrNoRows {
				if remember == "on" {
					remember = "checked"
				}
				renderTemplate(w, "signin.html",
					map[string]interface{}{
						"Email":         email,
						"FocusOnPasswd": true,
						"Remember":      remember,
						"ErrAuthFailed": true,
					})
			} else {
				httpError(w, 503, err)
			}
			return
		}
		// Create session and redirect to members area
		s, _ := Session.Get(r, "s")
		s.Values["Id"] = u.Id
		if remember == "on" {
			s.Options = &sessions.Options{Path: "/"}
		} else {
			s.Options = &sessions.Options{MaxAge: 0, Path: "/"}
		}
		s.Save(r, w)
		if next == "" {
			next = "/main/"
		}
		http.Redirect(w, r, next, 302)
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405, nil)
	}
}

func SignOutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "s", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", 302)
}

func MainHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
	if u, err := GetUserById(s.Values["Id"].(int)); err != nil {
		httpError(w, 503, err)
	} else {
		renderTemplate(w, "in/main.html",
			map[string]interface{}{"User": u, "MainMenu": true})
	}
}

func SettingsHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
	u, err := GetUserById(s.Values["Id"].(int))
	if err != nil {
		httpError(w, 503, err)
		return
	}
	switch r.Method {
	case "GET":
		renderTemplate(w, "in/settings.html",
			map[string]interface{}{"User": u})
	case "POST":
		changes := 0
		n := strings.Trim(r.FormValue("name"), "")
		nl := len(n)
		if nl >= 1 && n != u.FullName.String {
			if nl > 80 {
				renderTemplate(w, "in/settings.html",
					map[string]bool{"ErrInvalidName": true})
				return
			}
			u.FullName.String = n
			changes++
		}
		p1 := r.FormValue("passwd1")
		if p1 != "" {
			if len(p1) < 4 {
				renderTemplate(w, "in/settings.html",
					map[string]interface{}{
						"User":             u,
						"FocusOnPasswd1":   true,
						"ErrInvalidPasswd": true,
					})
				return
			}
			if p1 != r.FormValue("passwd2") {
				renderTemplate(w, "in/settings.html",
					map[string]interface{}{
						"User":              u,
						"FocusOnPasswd1":    true,
						"ErrPasswdMismatch": true,
					})
				return
			}
			h := sha1.New()
			io.WriteString(h, r.FormValue("passwd0")) // old pwd
			if hex.EncodeToString(h.Sum(nil)) != u.Passwd {
				renderTemplate(w, "in/settings.html",
					map[string]interface{}{
						"User":           u,
						"FocusOnPasswd0": true,
						"ErrAuthFailed":  true,
					})
				return
			}
			u.Passwd = p1
			changes++
		}
		var saved bool
		if changes >= 1 {
			if err := UpdateUser(u); err != nil {
				httpError(w, 503, err)
				return
			}
			saved = true
		}
		renderTemplate(w, "in/settings.html",
			map[string]interface{}{"User": u, "Saved": saved})
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405, nil)
	}
}
