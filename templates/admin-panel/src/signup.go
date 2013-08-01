// Copyright 2013 %template% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

type SignupRequest struct {
	Email      string
	InviteCode string
	TOS        bool
}

type SignupResponse struct {
	Ok    bool
	Error string
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Response Ok is the server is invite only
		JSON(w, SignupResponse{Ok: Config.InviteOnly})
		return
	case "POST":
		// Ignore, handle later.
	default:
		w.Header().Set("Allow", "GET, POST")
		httpError(w, 405)
	}
	// Handle POST
	var v SignupRequest
	err := ParseJSON(r, &v)
	if err != nil {
		httpError(w, 400, err)
		return
	}
	if Config.InviteOnly {
		if code, _ := Redis.Get("invite"); v.InviteCode != code {
			JSON(w, SignupResponse{Error: "InvalidInviteCode"})
			return
		}
	}
	if len(strings.Split(v.Email, "@")) != 2 {
		JSON(w, SignupResponse{Error: "InvalidEmail"})
		return
	}
	if !v.TOS {
		JSON(w, SignupResponse{Error: "TOS"})
		return
	}
	// Check if this user isn't pending confirmation.
	if exists, err := Redis.Exists("signup:" + v.Email); err != nil {
		httpError(w, 503, err)
		return
	} else if exists {
		JSON(w, SignupResponse{Error: "AwaitingConfirmation"})
		return
	}
	// Check if this user already exists in the db.
	if exists, err := UserExists(v.Email); err != nil {
		httpError(w, 503, err)
		return
	} else if exists {
		JSON(w, SignupResponse{Error: "AlreadyExists"})
		return
	}
	// Create random confirmation key and store in redis.
	hex := RandHex(24)
	if err := Redis.MSet(map[string]string{
		"signup:" + hex:     v.Email,
		"signup:" + v.Email: hex,
	}); err != nil {
		httpError(w, 503, err)
		return
	}
	Redis.Expire("signup:"+hex, 86400)
	Redis.Expire("signup:"+v.Email, 86400)
	// Render confirmation email message with the confirmation URL.
	msg := bytes.NewBuffer(nil)
	err = Tmpl.ExecuteTemplate(msg, "signup-email.txt",
		map[string]string{
			"ReplyTo": Config.SMTP.ReplyTo,
			"Email":   v.Email,
			"IP":      remoteIP(r),
			"URL":     serverURL(r, true) + "signup-confirm?q=" + hex,
		})
	if err != nil {
		Redis.Del("signup:"+hex, "signup:"+v.Email)
		httpError(w, 500, err)
		return
	}
	if err := SendMail([]string{v.Email}, msg.Bytes()); err != nil {
		// TODO: Check the error code and respond with
		//       different error pages.
		Redis.Del("signup:"+hex, "signup:"+v.Email)
		httpError(w, 503, err)
		return
	}
	JSON(w, SignupResponse{Ok: true})
}

type SignupConfirmRequest struct {
	URL      string
	FullName string
	Passwd   string
	Confirm  string
}

type SignupConfirmResponse struct {
	Ok    bool
	Error string
}

func SignupConfirmHandler(w http.ResponseWriter, r *http.Request) {
	var v SignupConfirmRequest
	err := ParseJSON(r, &v)
	if err != nil {
		httpError(w, 400, err)
		return
	}
	tmp := strings.Split(v.URL, "?q=")
	if len(tmp) != 2 {
		httpError(w, 404, nil)
		return
	}
	q := tmp[1]
	if ok, err := Redis.Exists(fmt.Sprintf("signup:%s", q)); err != nil {
		httpError(w, 503, err)
		return
	} else if !ok {
		httpError(w, 404, nil)
		return
	}
	if len(v.FullName) > 80 {
		JSON(w, SignupConfirmResponse{Error: "InvalidName"})
		return
	}
	if len(v.Passwd) < 4 {
		JSON(w, SignupConfirmResponse{Error: "InvalidPasswd"})
		return
	}
	if v.Passwd != v.Confirm {
		JSON(w, SignupConfirmResponse{Error: "PasswdMismatch"})
		return
	}
	// Get the email address from redis
	email, err := Redis.Get("signup:" + q)
	if err != nil {
		httpError(w, 503, err)
		return
	}
	// Create user in the db, activated.
	u, err := NewUser(email, v.Passwd, v.FullName, true)
	if err != nil {
		// Look for MySQL #1062 (dup entry)
		if strings.Contains(err.Error(), "#1062") {
			JSON(w, SignupConfirmResponse{Error: "AlreadyExists"})
		} else {
			httpError(w, 503, err)
		}
		return
	}
	// Delete sign up confirmation keys from redis.
	Redis.Del("signup:"+q, "signup:"+email)
	// Create session and return Ok to redirect to members area.
	s, _ := Session.Get(r, "s")
	s.Options = &sessions.Options{MaxAge: 0, Path: "/"}
	s.Values["Id"] = u.Id
	s.Save(r, w)
	// Render welcome email message with the system's URL.
	msg := bytes.NewBuffer(nil)
	err = Tmpl.ExecuteTemplate(msg, "signup-confirm-email.txt",
		map[string]string{
			"ReplyTo": Config.SMTP.ReplyTo,
			"Email":   email,
			"IP":      remoteIP(r),
			"URL":     serverURL(r, true),
		})
	if err != nil {
		httpError(w, 500, err)
		return
	}
	SendMail([]string{email}, msg.Bytes())
	JSON(w, SignupConfirmResponse{Ok: true})
}
