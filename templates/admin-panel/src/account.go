// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// HTTP handlers related to the user account (signup, login, passwd recovery).

package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := Session.Get(r, "s")
	if r.URL.Path == "/" && s != nil && s.Values["Id"] != nil {
		// Logged in. What now?
		fmt.Fprintf(w, "sup, yo?")
	} else {
		DocumentRoot.ServeHTTP(w, r)
	}
}

type SignupData struct {
	Title                string
	InviteOnly           bool
	Email                string
	InviteCode           string
	TOS                  string
	InvalidEmail         bool
	UserExists           bool
	AwaitingConfirmation bool
	InvalidInviteCode    bool
	UncheckedTOS         bool
}

type AccountEmailData struct {
	ReplyTo string
	Email   string
	IP      string
	URL     string
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		SignupTmpl.ExecuteTemplate(w, "base", &SignupData{
			Title:      "Sign up",
			InviteOnly: cfg.InviteOnly,
		})
		return
	case "POST":
		// Handle POST below.
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405)
	}

	// Handle POST
	d := &SignupData{
		Title:      "Sign up",
		InviteOnly: cfg.InviteOnly,
		Email:      r.FormValue("Email"),
		InviteCode: r.FormValue("InviteCode"),
		TOS:        r.FormValue("TOS"),
	}

	if len(strings.Split(d.Email, "@")) != 2 || len(d.Email) > 64 {
		d.InvalidEmail = true
		SignupTmpl.ExecuteTemplate(w, "base", d)
		return
	}

	if cfg.InviteOnly {
		if code, _ := Redis.Get("InviteCode"); d.InviteCode != code {
			d.InvalidInviteCode = true
			SignupTmpl.ExecuteTemplate(w, "base", d)
			return
		}
	}

	if d.TOS != "on" {
		d.UncheckedTOS = true
		d.TOS = ""
		SignupTmpl.ExecuteTemplate(w, "base", d)
		return
	} else {
		d.TOS = "checked"
	}

	var (
		exists bool
		err    error
	)

	// Check if the email is pending confirmation.
	if exists, err = Redis.Exists("Signup:" + d.Email); err != nil {
		httpError(w, 503, err)
		return
	} else if exists {
		d.AwaitingConfirmation = true
		SignupTmpl.ExecuteTemplate(w, "base", d)
		return
	}

	// Check if this email already exists in the db.
	if exists, err = UserExists(d.Email); err != nil {
		httpError(w, 503, err)
		return
	} else if exists {
		d.UserExists = true
		SignupTmpl.ExecuteTemplate(w, "base", d)
		return
	}

	// Create random confirmation key and store in redis.
	hex := RandHex(24)
	if err = Redis.MSet(map[string]string{
		"Signup:" + hex:     d.Email,
		"Signup:" + d.Email: hex,
	}); err != nil {
		httpError(w, 503, err)
		return
	}
	Redis.Expire("Signup:"+hex, 86400)
	Redis.Expire("Signup:"+d.Email, 86400)

	// Render confirmation email message with the confirmation URL.
	msg := bytes.NewBuffer(nil)
	if err = EmailTmpl.ExecuteTemplate(msg, "signup", &AccountEmailData{
		Email:   d.Email,
		ReplyTo: cfg.Email.ReplyTo,
		IP:      remoteIP(r),
		URL:     serverURL(r, true) + "hello/?q=" + hex,
	}); err != nil {
		Redis.Del("Signup:"+hex, "Signup:"+d.Email)
		httpError(w, 500, err)
		return
	}

	SendMail(msg.Bytes(), d.Email)
	SignupOkTmpl.ExecuteTemplate(w, "base",
		struct{ Title string }{"Sign up"})
}

type HelloData struct {
	Title          string
	Email          string
	FullName       string
	Passwd         string
	Passwd2        string
	AlreadyActive  bool
	InvalidName    bool
	InvalidPasswd  bool
	PasswdMismatch bool
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	if len(q) == 0 {
		httpError(w, 404)
		return
	}
	email, err := Redis.Get(fmt.Sprintf("Signup:%s", q))
	if err != nil {
		httpError(w, 503, "Redis: "+err.Error())
		return
	} else if len(email) == 0 {
		httpError(w, 404)
		return
	}

	switch r.Method {
	case "GET":
		HelloTmpl.ExecuteTemplate(w, "base", &HelloData{
			Title: "Sign up confirmation",
			Email: email,
		})
		return
	case "POST":
		// Handle POST below.
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405)
	}

	// Handle POST
	d := &HelloData{
		Title:    "Sign up confirmation",
		Email:    email,
		FullName: r.FormValue("FullName"),
		Passwd:   r.FormValue("Passwd"),
		Passwd2:  r.FormValue("Passwd2"),
	}

	if len(d.FullName) > 80 {
		d.InvalidName = true
		HelloTmpl.ExecuteTemplate(w, "base", d)
		return
	}

	if len(d.Passwd) < 4 {
		d.InvalidPasswd = true
		HelloTmpl.ExecuteTemplate(w, "base", d)
		return
	}

	if d.Passwd != d.Passwd2 {
		d.PasswdMismatch = true
		HelloTmpl.ExecuteTemplate(w, "base", d)
		return
	}

	// Create user in the db, activated.
	if _, err = NewUser(email, d.Passwd, d.FullName, true); err != nil {
		// Look for MySQL #1062 (dup entry)
		if strings.Contains(err.Error(), "#1062") {
			d.AlreadyActive = true
			HelloTmpl.ExecuteTemplate(w, "base", d)
		} else {
			httpError(w, 503, err)
		}
		return
	}

	// Delete sign up confirmation keys from redis.
	Redis.Del("Signup:"+q, "Signup:"+email)

	// Render and send welcome email message.
	msg := bytes.NewBuffer(nil)
	if err = EmailTmpl.ExecuteTemplate(msg, "hello", &AccountEmailData{
		Email:   d.Email,
		ReplyTo: cfg.Email.ReplyTo,
		IP:      remoteIP(r),
		URL:     serverURL(r, true),
	}); err != nil {
		httpError(w, 500, err)
		return
	}
	SendMail(msg.Bytes(), d.Email)

	// Redirect to login page.
	http.Redirect(w, r, serverURL(r, true)+"login/", 307)
}

type LoginData struct {
	Title      string
	Email      string
	Passwd     string
	RememberMe string
	Failed     bool
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		LoginTmpl.ExecuteTemplate(w, "base", &LoginData{
			Title: "Sign in",
		})
		return
	case "POST":
		// Handle POST below.
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405)
	}

	// Handle POST
	d := &LoginData{
		Title:      "Sign in",
		Email:      r.FormValue("Email"),
		Passwd:     r.FormValue("Passwd"),
		RememberMe: r.FormValue("RememberMe"),
	}

	if d.RememberMe == "on" {
		d.RememberMe = "checked"
	}

	u, err := GetUserWithPasswd(d.Email, d.Passwd)
	if err != nil {
		if err == sql.ErrNoRows {
			d.Failed = true
			LoginTmpl.ExecuteTemplate(w, "base", d)
			return
		} else {
			httpError(w, 503, err)
		}
		return
	}

	// Create session to allow access to members area.
	s, _ := Session.Get(r, "s")
	s.Values["Id"] = u.Id
	if len(d.RememberMe) > 3 {
		s.Options = &sessions.Options{Path: "/"}
	} else {
		s.Options = &sessions.Options{MaxAge: 0, Path: "/"}
	}
	s.Save(r, w)

	http.Redirect(w, r, "../", 307)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "s", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "../", 302)
}

func RecoveryHandler(w http.ResponseWriter, r *http.Request) {
}
