// Copyright 2013 %template% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

type UserIndexResponse struct {
	Ok       bool
	Email    string
	FullName string
}

func UserIndexHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
	if u, err := GetUserById(s.Values["Id"].(int)); err != nil {
		httpError(w, 503, err)
	} else {
		JSON(w, UserIndexResponse{
			Ok:       true,
			Email:    u.Email,
			FullName: u.FullName.String,
		})
	}
}

type UserSettingsRequest struct {
	FullName  string
	OldPasswd string
	NewPasswd string
	Confirm   string
}

type UserSettingsResponse struct {
	Ok      bool
	Changes int
	Error   string
}

func UserSettingsHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
	if r.Method == "GET" {
		UserIndexHandler(w, r, s)
		return
	}
	var v UserSettingsRequest
	err := ParseJSON(r, &v)
	if err != nil {
		httpError(w, 400, err)
		return
	}
	u, err := GetUserById(s.Values["Id"].(int))
	if err != nil {
		httpError(w, 503, err)
		return
	}
	changes := 0
	n := strings.Trim(v.FullName, "")
	nl := len(n)
	if nl > 0 && n != u.FullName.String {
		if nl > 80 {
			JSON(w, UserSettingsResponse{Error: "InvalidName"})
			return
		}
		u.FullName.String = n
		changes++
	}
	if v.NewPasswd != "" {
		if len(v.NewPasswd) < 4 {
			JSON(w, UserSettingsResponse{Error: "InvalidPasswd"})
			return
		}
		if v.NewPasswd != v.Confirm {
			JSON(w, UserSettingsResponse{Error: "PasswdMismatch"})
			return
		}
		h := sha1.New()
		io.WriteString(h, v.OldPasswd) // old pwd
		if hex.EncodeToString(h.Sum(nil)) != u.Passwd {
			JSON(w, UserSettingsResponse{Error: "InvalidOldPasswd"})
			return
		}
		newpw := sha1.New()
		io.WriteString(newpw, v.NewPasswd)
		u.Passwd = hex.EncodeToString(newpw.Sum(nil))
		changes++
	}
	if changes > 0 {
		if err := UpdateUser(u); err != nil {
			httpError(w, 503, err)
			return
		}
	}
	JSON(w, UserSettingsResponse{Ok: true, Changes: changes})
}
