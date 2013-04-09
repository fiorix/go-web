// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/sessions"
)

func RecoveryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "recovery.html", nil)
	case "POST":
		email := r.FormValue("email")
		if len(strings.Split(email, "@")) != 2 {
			renderTemplate(w, "signup.html",
				map[string]interface{}{
					"Email":           email,
					"ErrInvalidEmail": true,
				})
			return
		}
		if exists, err := UserExists(email); err != nil {
			httpError(w, 503, err)
			return
		} else if !exists {
			renderTemplate(w, "recovery.html",
				map[string]interface{}{
					"Email":          email,
					"ErrUnknownUser": true,
				})
			return
		}
		// Check if this user isn't pending confirmation.
		if exists, err := Redis.Exists("recovery:" + email); err != nil {
			httpError(w, 503, err)
			return
		} else if exists {
			renderTemplate(w, "recovery.html",
				map[string]bool{"ErrConfirmation": true})
			return
		}
		// Create random confirmation key and store in redis.
		hex := RandHex(24)
		if err := Redis.MSet(map[string]string{
			"recovery:" + hex:   email,
			"recovery:" + email: hex,
		}); err != nil {
			httpError(w, 503, err)
			return
		}
		Redis.Expire("recovery:"+hex, 86400)
		Redis.Expire("recovery:"+email, 86400)
		// Render recovery email message with the confirmation URL.
		url := fmt.Sprintf(
			"http://%s/recovery/confirm/?q=%s", r.Host, hex)
		msg, err := renderTemplateBytes("email/recovery.txt",
			map[string]string{
				"ReplyTo": Config.SMTP.ReplyTo,
				"Email":   email,
				"IP":      r.RemoteAddr,
				"URL":     url,
			})
		if err != nil {
			Redis.Del("recovery:"+hex, "recovery:"+email)
			httpError(w, 500, err)
			return
		}
		if err := sendMail([]string{email}, msg); err != nil {
			// TODO: Check the error code and respond with
			//       different error pages.
			Redis.Del("recovery:"+hex, "recovery:"+email)
			httpError(w, 503, err)
			return
		}
		renderTemplate(w, "recovery_ok.html",
			map[string]string{"Email": email})
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405, nil)
	}
}

func RecoveryConfirmHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	if q == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	if ok, err := Redis.Exists(fmt.Sprintf("recovery:%s", q)); err != nil {
		httpError(w, 503, err)
		return
	} else if !ok {
		httpError(w, 404, nil)
		return
	}
	switch r.Method {
	case "GET":
		renderTemplate(w, "recovery_confirm.html", nil)
	case "POST":
		// p := strings.Trim(r.FormValue("passwd"), " ")
		p := r.FormValue("passwd") // don't care about spaces
		if len(p) < 4 {
			renderTemplate(w, "recovery_confirm.html",
				map[string]interface{}{
					"ErrInvalidPasswd": true,
				})
			return
		}
		if p != r.FormValue("confirm") {
			renderTemplate(w, "recovery_confirm.html",
				map[string]interface{}{
					"ErrPasswdMismatch": true,
				})
			return
		}
		// Get the email address from redis
		email, err := Redis.Get("recovery:" + q)
		if err != nil {
			httpError(w, 503, err)
			return
		}
		// Get user from DB
		u, err := GetUser(email)
		if err != nil {
			// TODO: check sql.ErrNoRows
			httpError(w, 503, err)
			return
		}
		u.Passwd = p
		if err := UpdateUser(u); err != nil {
			httpError(w, 503, err)
			return
		}
		// Delete recovery confirmation keys from redis.
		Redis.Del("recovery:"+q, "recovery:"+email)
		// Create session and redirect to members area.
		s, _ := Session.Get(r, "s")
		s.Options = &sessions.Options{MaxAge: 0, Path: "/"}
		s.Values["Id"] = u.Id
		s.Save(r, w)
		http.Redirect(w, r, "/main", 302)
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405, nil)
	}
}
