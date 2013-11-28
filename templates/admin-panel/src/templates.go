// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"html/template"
	"path/filepath"
	text_template "text/template"
)

var (
	EmailTmpl    *text_template.Template
	SignupTmpl   *template.Template
	SignupOkTmpl *template.Template
	HelloTmpl    *template.Template
	LoginTmpl    *template.Template
)

func LoadTemplates() {
	EmailTmpl = text_template.Must(text_template.New("Email").ParseGlob(
		filepath.Join(cfg.TemplatesDir, "email", "*.txt"),
	))

	SignupTmpl = template.Must(template.New("Signup").ParseFiles(
		filepath.Join(cfg.TemplatesDir, "account", "base.html"),
		filepath.Join(cfg.TemplatesDir, "account", "signup.html"),
	))

	SignupOkTmpl = template.Must(template.New("SignupOk").ParseFiles(
		filepath.Join(cfg.TemplatesDir, "account", "base.html"),
		filepath.Join(cfg.TemplatesDir, "account", "signup_ok.html"),
	))

	HelloTmpl = template.Must(template.New("Hello").ParseFiles(
		filepath.Join(cfg.TemplatesDir, "account", "base.html"),
		filepath.Join(cfg.TemplatesDir, "account", "hello.html"),
	))

	LoginTmpl = template.Must(template.New("Login").ParseFiles(
		filepath.Join(cfg.TemplatesDir, "account", "base.html"),
		filepath.Join(cfg.TemplatesDir, "account", "login.html"),
	))
}
