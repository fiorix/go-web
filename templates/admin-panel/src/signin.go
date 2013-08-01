// Copyright 2013 %template% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
)

type SigninRequest struct {
	Email    string
	Passwd   string
	Remember bool
}

type SigninResponse struct {
	Ok    bool
	Error string
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	var v SigninRequest
	err := ParseJSON(r, &v)
	if err != nil {
		httpError(w, 400, err)
		return
	}
	if len(strings.Split(v.Email, "@")) != 2 {
		JSON(w, SigninResponse{Error: "InvalidEmail"})
		return
	}
	u, err := GetUserWithPasswd(v.Email, v.Passwd)
	if err != nil {
		if err == sql.ErrNoRows {
			JSON(w, SigninResponse{Error: "InvalidEmail"})
		} else {
			httpError(w, 503, err)
		}
		return
	}
	// Create session to allow access to members area.
	s, _ := Session.Get(r, "s")
	s.Values["Id"] = u.Id
	if v.Remember {
		s.Options = &sessions.Options{Path: "/"}
	} else {
		s.Options = &sessions.Options{MaxAge: 0, Path: "/"}
	}
	s.Save(r, w)
	JSON(w, SigninResponse{Ok: true})
}

func SignoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "s", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", 302)
}
