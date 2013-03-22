// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/nuswit/go-web"
	"time"
)

func IndexHandler(req *web.RequestHandler) {
	fmt.Printf("URL regexp: %q\n", req.Vars)
	switch(req.HTTP.Method) {
	case "GET":
		req.Redirect("/login?next=" + req.HTTP.URL.Path)
	case "POST":
		// Form vars. Same thing for GET.
		foobar := req.HTTP.Form.Get("foobar")
		req.Write("Hello, %s", foobar)
	default:
		// HTTP 405 (Method Not Allowed) is always an option. A list
		// of allowed methods must be provided.
		// See http://www.ietf.org/rfc/rfc2616.txt #14.7 for details.
		allowed := []string{"GET", "POST"}
		req.NotAllowed(allowed)
	}
}

func FoobarHandler(req *web.RequestHandler) {
	req.Render("foobar.html", map[string]interface{}{
		"Foo": "bar", // must be capitalized
		"Items": []string{"a", "b", "c"},
	})
}

func main() {
	handlers := []web.Handler{
		{"^/(a|b|c)?$", IndexHandler},
		{"^/foobar/?$", FoobarHandler},
	}
	settings := web.Settings{
		Debug: true,
		XHeaders: true,
		TemplatePath: "./templates",
		ReadTimeout: 30*time.Second,
		WriteTimeout: 10*time.Second,
	}
	web.Application(":8080", handlers, &settings)
}
