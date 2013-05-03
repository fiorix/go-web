// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// remux is a simple request multiplexer that supports regular expressions.
//
// Example:
//
//    package main
//
//    import (
//    	"fmt"
//    	"net/http"
//
//    	"github.com/fiorix/go-web/remux"
//    )
//
//    func IndexHandler(w http.ResponseWriter, r *http.Request) {
//    	vars := remux.Vars(r)
//    	fmt.Fprintln(w, "Hello, world", vars)
//    }
//
//    func main() {
//    	remux.HandleFunc("^/(foo|bar)?$", IndexHandler)
//    	server := http.Server{
//    		Addr:    ":8080",
//    		Handler: remux.DefaultServeMux,
//    	}
//    	server.ListenAndServe()
//    }
package remux

import (
	"net/http"
	"path"
	"regexp"
	"sync"
)

// ServeMux is an HTTP request multiplexer.
// It matches the URL of each incoming request against a list of registered
// patterns and calls the handler for the pattern that
// most closely matches the URL.
//
// Patterns are regular expressions, like "^/$". On routing decision,
// the handler of the first regex that match against URL.Path is executed.
//
// Patterns may optionally begin with a host name, restricting matches to
// URLs on that host only.  Host-specific patterns take precedence over
// general patterns, so that a handler might register for the two patterns
// "/codesearch" and "codesearch.google.com/" without also taking over
// requests for "http://www.google.com/".
//
// ServeMux also takes care of sanitizing the URL request path,
// redirecting any request containing . or .. elements to an
// equivalent .- and ..-free URL.
type ServeMux struct {
	mu sync.RWMutex
	m  map[*regexp.Regexp]muxEntry
}

type muxEntry struct {
	explicit bool
	h        http.Handler
}

var vdata map[*http.Request][]string
var vlock sync.RWMutex

func setVar(r *http.Request, m []string) {
	if vdata == nil {
		vdata = make(map[*http.Request][]string)
	}
	vlock.Lock()
	defer vlock.Unlock()
	vdata[r] = m
}

func delVar(r *http.Request) {
	if vdata == nil {
		return
	}
	vlock.RLock()
	_, exists := vdata[r]
	vlock.RUnlock()
	if exists {
		vlock.Lock()
		defer vlock.Unlock()
		delete(vdata, r)
	}
}

// Vars returns the result of the regex execution on the URL pattern.
func Vars(r *http.Request) []string {
	if vdata == nil {
		return nil
	}
	vlock.RLock()
	defer vlock.RUnlock()
	return vdata[r]
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{m: make(map[*regexp.Regexp]muxEntry)}
}

// DefaultServeMux is the default ServeMux used by Serve.
var DefaultServeMux = NewServeMux()

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
func (mux *ServeMux) match(path string) ([]string, http.Handler) {
	for k, v := range mux.m {
		m := k.FindStringSubmatch(path)
		if len(m) >= 1 {
			return m[1:], v.h // m[0] is URL.Path thus not needed
		}
	}
	return nil, nil
}

// handler returns the handler to use for the request r.
func (mux *ServeMux) handler(r *http.Request) http.Handler {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	m, h := mux.match(r.Host + r.URL.Path)
	if h == nil {
		m, h = mux.match(r.URL.Path)
	}
	if h == nil {
		h = http.NotFoundHandler()
	}
	// Vars hold the result of the pattern regexp executed on URL.Path
	setVar(r, m)
	return h
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "CONNECT" {
		// Clean path to canonical form and redirect.
		if p := cleanPath(r.URL.Path); p != r.URL.Path {
			w.Header().Set("Location", p)
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
	}
	mux.handler(r).ServeHTTP(w, r)
	delVar(r)
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (mux *ServeMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("http: nil handler")
	}
	pattern_re := regexp.MustCompile(pattern)
	if mux.m[pattern_re].explicit {
		panic("http: multiple registrations for " + pattern)
	}

	mux.m[pattern_re] = muxEntry{explicit: true, h: handler}

	// Helpful behavior:
	// If pattern is /tree/, insert an implicit permanent redirect for /tree.
	// It can be overridden by an explicit registration.
	/*
		n := len(pattern)
		if n > 0 && pattern[n-1] == '/' && !mux.m[pattern[0:n-1]].explicit {
			mux.m[pattern[0:n-1]] = muxEntry{h: RedirectHandler(pattern, StatusMovedPermanently)}
		}
	*/
}

// HandleFunc registers the handler function for the given pattern.
func (mux *ServeMux) HandleFunc(pattern string,
	handler func(http.ResponseWriter, *http.Request)) {
	mux.Handle(pattern, http.HandlerFunc(handler))
}

// HandleFunc registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	DefaultServeMux.HandleFunc(pattern, handler)
}
