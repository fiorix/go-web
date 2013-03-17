// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package web

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
	"log"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type RequestHandler struct {
	Writer http.ResponseWriter
	HTTP *http.Request
	Server *Server
	Vars []string
}

func (req *RequestHandler) HTTPError(n int, e ...error) {
	if e != nil && req.Server.settings.Debug {
		log.Printf("HTTPError %d: %s", n, e[0])  // whine
	}
	http.Error(req.Writer, http.StatusText(n), n)
}

func (req *RequestHandler) NotFound() {
	http.NotFound(req.Writer, req.HTTP)
}

func (req *RequestHandler) Redirect(url string) {
	http.Redirect(req.Writer, req.HTTP, url, http.StatusFound)
}

func (req *RequestHandler) Render(t string, a interface{}) error {
	if req.Server.templates == nil {
		e := errors.New("TemplatePath not set in web.Settings")
		if req.Server.settings.Debug {
			log.Println(e)
		}
		return e
	}

	err := req.Server.templates.ExecuteTemplate(req.Writer, t, a)
	if err != nil && req.Server.settings.Debug {
		log.Println(err)
	}
	return err
}

func (req *RequestHandler) ServeFile(name string) {
	http.ServeFile(req.Writer, req.HTTP, name)
}

func (req *RequestHandler) SetHeader(k string, v string) {
	req.Writer.Header().Set(k, v)
}

func (req *RequestHandler) Write(f string, a ...interface{}) (int, error) {
	return fmt.Fprintf(req.Writer, f, a...)
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
	XHeaders bool
	TemplatePath string
	ReadTimeout time.Duration
	WriteTimeout time.Duration
}

type Server struct {
	routes []route
	settings *Settings
	templates *template.Template
}

func execute(fn func(RequestHandler), req RequestHandler) {
	var now time.Time
	if req.Server.settings.Debug {
		now = time.Now()
	}
	if req.Server.settings.XHeaders {
		addr := req.HTTP.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.HTTP.Header.Get("X-Forwarded-For")
		}

		if addr != "" {
			req.HTTP.RemoteAddr = addr
		}
	}
	fn(req)  // execute the HandlerFunc
	if req.Server.settings.Debug {
		ra := req.HTTP.RemoteAddr
		if ra == "" {
			ra = "unix"
		}
		log.Printf("%s %s (%s) %s",
				req.HTTP.Method,
				req.HTTP.URL.Path,
				ra, time.Since(now))
	}
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, p := range srv.routes {
		vars := p.re.FindStringSubmatch(r.URL.Path)
		if len(vars) >= 1 {
			execute(p.fn, RequestHandler{w, r, srv, vars})
			return
		}
	}
	http.NotFound(w, r)
}

// Support for unix sockets
func ListenAndServe(srv *http.Server) (net.Listener, error) {
	// code from http://golang.org/src/pkg/net/http/server.go
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	var proto string
	if strings.HasPrefix(addr, "unix:") {
		proto = "unix"
		addr = addr[5:]  // len("unix:")
		syscall.Unlink(addr)
	} else {
		proto = "tcp"
	}
	return net.Listen(proto, addr)
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
		log.Println("Starting server on", addr)
	}
	rtimeout := 0*time.Second  // Keep-alive might be your enemy here
	if s.ReadTimeout >= 1 {
		rtimeout = s.ReadTimeout
	}
	wtimeout := 0*time.Second
	if s.WriteTimeout >= 1 {
		wtimeout = s.WriteTimeout
	}
	ws := Server{r, s, t}
	srv := &http.Server{Addr: addr, Handler: &ws,
				ReadTimeout: rtimeout, WriteTimeout:wtimeout}
	// e := srv.ListenAndServe()
	l, e := ListenAndServe(srv)
	if e != nil {
		if s.Debug {
			log.Println("Error:", e)
		}
		return nil, e
	}

	return &ws, srv.Serve(l)
}
