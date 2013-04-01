// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// autogzip provides on-the-fly gzip encoding for http servers.

package autogzip

import (
	"compress/gzip"
	"github.com/fiorix/go-web/http"
	"io"
	"strings"
)

type ResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Handle provides on-the-fly gzip encoding for other handlers.
//
// Usage:
//
//	func IndexHandler(w http.ResponseWriter, req *http.Request) {
//		fmt.Fprintln(w, "Hello, world")
//	}
//
//	func main() {
//		http.HandleFunc("/", IndexHandler)
//		http.ListenAndServe(":8080", autogzip.Handle(http.DefaultServeMux))
//	}
func Handle(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		h.ServeHTTP(ResponseWriter{Writer: gz, ResponseWriter: w}, r)
	}
}
