/* Copyright 2013 Alexandre Fiori
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

package web

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"
)

type RequestHandler struct {
	Writer http.ResponseWriter
	HTTP *http.Request
	Server *Server
	Vars []string
}

func (req *RequestHandler) Write(f string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(req.Writer, f, a)
}

func (req *RequestHandler) Render(t string, a interface{}) error {
	if req.Server.templates == nil {
		e := errors.New("TemplatePath not set in web.Settings")
		if req.Server.settings.Debug {
			fmt.Println(e)
		}
		return e
	}

	err := req.Server.templates.ExecuteTemplate(req.Writer, t, a)
	if err != nil && req.Server.settings.Debug {
		fmt.Println(err)
	}
	return err
}

func (req *RequestHandler) Redirect(url string) {
	http.Redirect(req.Writer, req.HTTP, url, http.StatusFound)
}

func (req *RequestHandler) NotFound() {
	http.NotFound(req.Writer, req.HTTP)
}

func (req *RequestHandler) HTTPError(n int, err error) {
	http.Error(req.Writer, err.Error(), n)
}

type HandlerFunc func(req RequestHandler)
type Handler struct {
	Re string
	Fn HandlerFunc
}

type route struct {
	re *regexp.Regexp
	fn HandlerFunc
}

type Settings struct {
	Debug bool
	TemplatePath string
}

type Server struct {
	routes []route
	settings *Settings
	templates *template.Template
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, p := range srv.routes {
		vars := p.re.FindStringSubmatch(r.URL.Path)
		if len(vars) >= 1 {
			p.fn(RequestHandler{w, r, srv, vars})
			return
		}
	}
	http.NotFound(w, r)
}

func Application(addr string, h []Handler, s *Settings) (*Server, error) {
	var t *template.Template
	if s.TemplatePath != "" {
		path := filepath.Join(s.TemplatePath, "*.html")
		t = template.Must(template.ParseGlob(path))
	}

	r := make([]route, len(h))
	for n, handler := range h {
		r[n] = route{regexp.MustCompile(handler.Re), handler.Fn}
	}

	if s.Debug {
		fmt.Println("Starting server on", addr)
	}

	srv := Server{r, s, t}
	return &srv, http.ListenAndServe(addr, &srv)
}
