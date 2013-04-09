// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/remux"
)

const (
	pagesDir    = "./pages"
	pagesDirLen = len(pagesDir) - 1
	extLen      = len(".txt")
)

var templates = template.Must(template.ParseGlob("./templates/*.html"))

func renderTemplate(w http.ResponseWriter, name string, a interface{}) error {
	err := templates.ExecuteTemplate(w, name, a)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
	return err
}

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := filepath.Join(pagesDir, p.Title+".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := filepath.Join(pagesDir, title+".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	files, err := filepath.Glob(filepath.Join(pagesDir, "*.txt"))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	type PageList struct{ Title string }
	pages := make([]PageList, len(files))
	for n, name := range files {
		v := name[pagesDirLen:]
		pages[n].Title = v[:len(v)-extLen]
	}
	renderTemplate(w, "index.html", map[string]interface{}{"Pages": pages})
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.Vars[0]
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.Vars[0]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit.html", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.Vars[0]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func logger(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP %d %s %s (%s) :: %s",
		w.Status(),
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		time.Since(r.Created))
}

func main() {
	title_re := "([a-zA-Z0-9]+)$"
	remux.HandleFunc("^/$", IndexHandler)
	remux.HandleFunc("^/view/"+title_re, viewHandler)
	remux.HandleFunc("^/edit/"+title_re, editHandler)
	remux.HandleFunc("^/save/"+title_re, saveHandler)
	server := http.Server{
		Addr:    ":8080",
		Handler: remux.DefaultServeMux,
		Logger:  logger,
	}
	server.ListenAndServe()
}
