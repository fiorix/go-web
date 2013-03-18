// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This code is very experimental at this point and names might change,
// things disappear, etc.
package web

import (
	"bufio"
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

// This is used to interact with the HTTP request and response, render
// templates, serve events (SSE), etc.
//
// func IndexHandler(req web.RequestHandler) {
//   ...
// }
type RequestHandler struct {
	// The response writer, used to set headers and write data
	// back to the client.
	Writer http.ResponseWriter

	// Request information: Method, URL, etc
	HTTP *http.Request
	Server *Server

	// Result of the handlers's regexp executed on the URL:
	// "^/(a|b|c)/$", IndexHandler...
	Vars []string
}

// Returns an HTTP error to the client. The optional log message is only
// printed when the server is in debug mode.
func (req *RequestHandler) HTTPError(n int, f string, a ...interface{}) {
	if f != "" && req.Server.Settings.Debug {
		log.Printf(f, a...)
	}
	http.Error(req.Writer, http.StatusText(n), n)
}

// Returns HTTP 404
func (req *RequestHandler) NotFound() {
	http.NotFound(req.Writer, req.HTTP)
}

// Returns HTTP 302 and with Location header set to "url"
func (req *RequestHandler) Redirect(url string) {
	http.Redirect(req.Writer, req.HTTP, url, http.StatusFound)
}

// Renders the template "t" and writes the result to the client.
// Example:
//
// func IndexHandler(req web.RequestHandler) {
//   req.Render("index.html", map[string]interface{}{"foo": "bar"})
// }
//
// func main() {
//   web.Application(":8080", []web.Handler{{"^/$", IndexHandler}},
//                   &web.Settings{Debug:true, Template_path:"./templates"})
// }
//
// If "Template_path" is not provided during initialization, the first call to
// "Render" panics. (just like you)
//
// A compilation error is returned if the template fails to render. The
// error message is printed if the server is in debug mode.
func (req *RequestHandler) Render(t string, a interface{}) error {
	if req.Server.templates == nil {
		log.Println("req.Render(%s) failed.", t)
		panic("TemplatePath not set in web.Settings")
	}

	err := req.Server.templates.ExecuteTemplate(req.Writer, t, a)
	if err != nil && req.Server.Settings.Debug {
		log.Println(err)
	}
	return err
}

// Serves the file or directory "name".
// Use with caution, it might expose the entire filesystem. (../../etc)
func (req *RequestHandler) ServeFile(name string) {
	http.ServeFile(req.Writer, req.HTTP, name)
}

// Sets header "k" = "v":
func (req *RequestHandler) SetHeader(k string, v string) {
	req.Writer.Header().Set(k, v)
}

// An event message. (named after the spec)
// http://dev.w3.org/html5/eventsource/
//
// This is part of the Server-Sent Events implementation. Before sending
// messages the server must be in events mode, by calling req.ServeEvents().
// Example:
//
// func IndexHandler(req web.RequestHandler) {
//   conn, bufrw, err := req.ServeEvents()
//   if err != nil {
//     return
//   }
//   defer conn.Close()
//   for {
//     req.SendEvent(bufrw, &web.MessageEvent{Id:"foo", Data:"bar"})
//     ...
//   }
// }
type MessageEvent struct {
	Event string
	Data string
	Id string  // int?
	Retry int
}

// Sends an event. The server must be in events mode.
func (req *RequestHandler) SendEvent(bufrw *bufio.ReadWriter, m *MessageEvent) error {
	if m.Data != "" {
		fmt.Fprintf(bufrw, "data: %s\n", m.Data)
	}
	if m.Event != "" {
		fmt.Fprintf(bufrw, "event: %s\n", m.Event)
	}
	if m.Id != "" {
		fmt.Fprintf(bufrw, "id: %s\n", m.Id)
	}
	if m.Retry >= 1{
		fmt.Fprintf(bufrw, "retry: %d\n", m.Retry)
	}
	fmt.Fprintf(bufrw, "\n")
	return bufrw.Flush()
}

var NoHijack = errors.New("webserver doesn't support hijacking")

// Hijacks the HTTP client socket and returns it.
// It gives up the control over the request, therefore methods like
// .Write() and .SetHeader() no longer work. All other information like the
// request headers remain intact in req.HTTP(.URL, .Header, etc).
// Must be called only once, and puts the server in events mode - only
// the current request.
func (req *RequestHandler) ServeEvents() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := req.Writer.(http.Hijacker)
	if !ok {
		return nil, nil, NoHijack
	}
	conn, bufrw, err := hj.Hijack()
	if err == nil {
		fmt.Fprintf(bufrw,
			"HTTP/1.1 200 OK\r\n"+
			"Cache-Control: no-cache\r\n"+
			"Connection: keep-alive\r\n"+
			"Content-Type: text/event-stream\r\n\r\n")
		bufrw.Flush()
	}
	return conn, bufrw, err
}

// Writes data to the client. Uses the default transfer encoding, chunked.
func (req *RequestHandler) Write(f string, a ...interface{}) (int, error) {
	return fmt.Fprintf(req.Writer, f, a...)
}

type HandlerFunc func(req RequestHandler)

// Maps URI patterns to request handler (HandlerFunc) functions.
// Example:
//
// handlers := []web.Handler{
//   {"^/$": IndexHandler},
//   {"^/(a|b|c)/$": AbcHandler},
// }
type Handler struct {
	Re string  // Regexp for the URL. e.g.: ^/index.html$
	Fn HandlerFunc  // Handler function.
}

type route struct {
	re *regexp.Regexp
	fn HandlerFunc
}

// Settings used to initialize the server.
// Example:
//
// web.Application(":8080", handlers, &web.Settings{Debug:false, XHeaders:true})
type Settings struct {
	Debug bool  // Makes the entire server very noisy when set to true
	XHeaders bool  // Uses X-Real-IP or X-Forwarded-For HTTP headers when available
	TemplatePath string  // Initializes HTML templates in a directory
	ReadTimeout time.Duration  // Get rid of non-active keep-alive clients
	WriteTimeout time.Duration
}

// Base server. Might support methods like .AddHandler() and others in the
// future. Or not.
type Server struct {
	routes []route
	templates *template.Template
	Settings *Settings
}

func execute(fn func(RequestHandler), req RequestHandler) {
	var now time.Time
	if req.Server.Settings.Debug {
		now = time.Now()
	}
	if req.Server.Settings.XHeaders {
		addr := req.HTTP.Header.Get("X-Real-IP")
		if addr == "" {
			addr = req.HTTP.Header.Get("X-Forwarded-For")
		}

		if addr != "" {
			req.HTTP.RemoteAddr = addr
		}
	}
	fn(req)  // execute the HandlerFunc
	if req.Server.Settings.Debug {
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

// Executes a request handler. The handler is selected if its pattern regexp
// match the URL.Path. HTTP 404 is returned otherwise.
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, p := range srv.routes {
		vars := p.re.FindStringSubmatch(r.URL.Path)
		if len(vars) >= 1 {
			execute(p.fn, RequestHandler{
					Writer: w,
					HTTP: r,
					Server: srv,
					Vars: vars})
			return
		}
	}
	http.NotFound(w, r)
}

// Listens on ip:port or unix:/filename.
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

// Starts the application.
// Example:
//
// func IndexHandler(req web.RequestHandler) {
//   req.Write("Hello, world")
// }
//
// func main() {
//   web.Application(":8080", []web.Handler{{"^/$", IndexHandler}},
//                   &web.Settings{Debug:true})
// }
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
	ws := Server{r, t, s}
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
