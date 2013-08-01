// Copyright 2013 %template% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

type RecoveryRequest struct {
	Email string
}

type RecoveryResponse struct {
	Ok    bool
	Error string
}

func RecoveryHandler(w http.ResponseWriter, r *http.Request) {
	var v RecoveryRequest
	err := ParseJSON(r, &v)
	if err != nil {
		httpError(w, 400, err)
		return
	}
	if len(strings.Split(v.Email, "@")) != 2 {
		JSON(w, RecoveryResponse{Error: "InvalidEmail"})
		return
	}
	if exists, err := UserExists(v.Email); err != nil {
		httpError(w, 503, err)
		return
	} else if !exists {
		JSON(w, RecoveryResponse{Error: "DoesNotExist"})
		return
	}
	// Check if this user isn't pending confirmation.
	if exists, err := Redis.Exists("recovery:" + v.Email); err != nil {
		httpError(w, 503, err)
		return
	} else if exists {
		JSON(w, RecoveryResponse{Error: "AwaitingConfirmation"})
		return
	}
	// Create random confirmation key and store in redis.
	hex := RandHex(24)
	if err := Redis.MSet(map[string]string{
		"recovery:" + hex:     v.Email,
		"recovery:" + v.Email: hex,
	}); err != nil {
		httpError(w, 503, err)
		return
	}
	Redis.Expire("recovery:"+hex, 86400)
	Redis.Expire("recovery:"+v.Email, 86400)
	// Render recovery email message with the confirmation URL.
	msg := bytes.NewBuffer(nil)
	err = Tmpl.ExecuteTemplate(msg, "recovery-email.txt",
		map[string]string{
			"ReplyTo": Config.SMTP.ReplyTo,
			"Email":   v.Email,
			"IP":      remoteIP(r),
			"URL":     serverURL(r, true) + "recovery-confirm?q=" + hex,
		})
	if err != nil {
		Redis.Del("recovery:"+hex, "recovery:"+v.Email)
		httpError(w, 500, err)
		return
	}
	if err := SendMail([]string{v.Email}, msg.Bytes()); err != nil {
		// TODO: Check the error code and respond with
		//       different error pages.
		Redis.Del("recovery:"+hex, "recovery:"+v.Email)
		httpError(w, 503, err)
		return
	}
	JSON(w, RecoveryResponse{Ok: true})
}

type RecoveryConfirmRequest struct {
	URL     string
	Passwd  string
	Confirm string
}

type RecoveryConfirmResponse struct {
	Ok    bool
	Error string
}

func RecoveryConfirmHandler(w http.ResponseWriter, r *http.Request) {
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
	if ok, err := Redis.Exists(fmt.Sprintf("recovery:%s", q)); err != nil {
		httpError(w, 503, err)
		return
	} else if !ok {
		httpError(w, 404, nil)
		return
	}
	if len(v.Passwd) < 4 {
		JSON(w, RecoveryConfirmResponse{Error: "InvalidPasswd"})
		return
	}
	if v.Passwd != v.Confirm {
		JSON(w, RecoveryConfirmResponse{Error: "PasswdMismatch"})
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
	newpw := sha1.New()
	io.WriteString(newpw, v.Passwd)
	u.Passwd = hex.EncodeToString(newpw.Sum(nil))
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
	// Render confirmation email message with the system's URL.
	msg := bytes.NewBuffer(nil)
	err = Tmpl.ExecuteTemplate(msg, "recovery-confirm-email.txt",
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
	JSON(w, RecoveryConfirmResponse{Ok: true})
}
