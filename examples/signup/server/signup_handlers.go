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

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		renderTemplate(w, "signup.html", nil)
	case "POST":
		legal := r.FormValue("legal")
		email := r.FormValue("email")
		if len(strings.Split(email, "@")) != 2 {
			if legal == "on" {
				legal = "checked"
			}
			renderTemplate(w, "signup.html",
				map[string]interface{}{
					"Email":           email,
					"Legal":           legal,
					"ErrInvalidEmail": true,
				})
			return
		}
		if legal != "on" {
			renderTemplate(w, "signup.html",
				map[string]interface{}{
					"Email":  email,
					"ErrTOS": true,
				})
			return
		}
		// Check if this user isn't pending confirmation.
		if exists, err := Redis.Exists("signup:" + email); err != nil {
			httpError(w, 503, err)
			return
		} else if exists {
			renderTemplate(w, "signup.html",
				map[string]bool{"ErrConfirmation": true})
			return
		}
		// Check if this user already exists in the db.
		if exists, err := UserExists(email); err != nil {
			httpError(w, 503, err)
			return
		} else if exists {
			// Could send the username ?u=email
			http.Redirect(w, r, "/recovery/", 302)
			return
		}
		// Create random confirmation key and store in redis.
		hex := RandHex(24)
		if err := Redis.MSet(map[string]string{
			"signup:" + hex:   email,
			"signup:" + email: hex,
		}); err != nil {
			httpError(w, 503, err)
			return
		}
		Redis.Expire("signup:"+hex, 86400)
		Redis.Expire("signup:"+email, 86400)
		// Render welcome email message with the confirmation URL.
		url := fmt.Sprintf(
			"http://%s/signup/confirm/?q=%s", r.Host, hex)
		msg, err := renderTemplateBytes("email/signup.txt",
			map[string]string{
				"ReplyTo": Config.SMTP.ReplyTo,
				"Email":   email,
				"IP":      r.RemoteAddr,
				"URL":     url,
			})
		if err != nil {
			Redis.Del("signup:"+hex, "signup:"+email)
			httpError(w, 500, err)
			return
		}
		if err := sendMail([]string{email}, msg); err != nil {
			// TODO: Check the error code and respond with
			//       different error pages.
			Redis.Del("signup:"+hex, "signup:"+email)
			httpError(w, 503, err)
			return
		}
		renderTemplate(w, "signup_ok.html",
			map[string]string{"Email": email})
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405, nil)
	}
}

func SignUpConfirmHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	if q == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	if ok, err := Redis.Exists(fmt.Sprintf("signup:%s", q)); err != nil {
		httpError(w, 503, err)
		return
	} else if !ok {
		httpError(w, 404, nil)
		return
	}
	switch r.Method {
	case "GET":
		renderTemplate(w, "signup_confirm.html", nil)
	case "POST":
		n := r.FormValue("name")
		if len(n) > 80 {
			renderTemplate(w, "signup_confirm.html",
				map[string]bool{"ErrInvalidName": true})
			return
		}
		// p := strings.Trim(r.FormValue("passwd"), " ")
		p := r.FormValue("passwd") // don't care about spaces
		if len(p) < 4 {
			renderTemplate(w, "signup_confirm.html",
				map[string]interface{}{
					"Name":             n,
					"FocusOnPasswd":    true,
					"ErrInvalidPasswd": true,
				})
			return
		}
		if p != r.FormValue("confirm") {
			renderTemplate(w, "signup_confirm.html",
				map[string]interface{}{
					"Name":              n,
					"FocusOnPasswd":     true,
					"ErrPasswdMismatch": true,
				})
			return
		}
		// Get the email address from redis
		email, err := Redis.Get("signup:" + q)
		if err != nil {
			httpError(w, 503, err)
			return
		}
		// Create user in the db, activated.
		u, err := NewUser(email, p, n, true)
		if err != nil {
			// Look for MySQL #1062 (dup entry)
			if strings.Contains(err.Error(), "#1062") {
				http.Redirect(w, r, "/signin", 302)
			} else {
				httpError(w, 503, err)
			}
			return
		}
		// Delete sign up confirmation keys from redis.
		Redis.Del("signup:"+q, "signup:"+email)
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
