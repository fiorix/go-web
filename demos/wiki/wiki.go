// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"github.com/fiorix/web"
	"io/ioutil"
	"path/filepath"
)

const (
	pagesDir = "./pages"
	pagesDirLen = len(pagesDir)-1
	extLen = len(".txt")
)

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := filepath.Join(pagesDir, p.Title + ".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := filepath.Join(pagesDir, title + ".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func IndexHandler(req web.RequestHandler) {
	files, err := filepath.Glob(filepath.Join(pagesDir, "*.txt"))
	if err != nil {
		req.HTTPError(500, err)
		return
	}

	type PageList struct { Title string }
	pages := make([]PageList, len(files))
	for n, name := range files {
		v := name[pagesDirLen:]
		pages[n].Title = v[:len(v)-extLen]
	}
	req.Render("index.html", map[string]interface{} {"Pages": pages})
}

func viewHandler(req web.RequestHandler) {
	title := req.Vars[1]
	p, err := loadPage(title)
	if err != nil {
		req.Redirect("/edit/"+title)
		return
	}
	req.Render("view.html", p)
}

func editHandler(req web.RequestHandler) {
	title := req.Vars[1]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	req.Render("edit.html", p)
}

func saveHandler(req web.RequestHandler) {
	title := req.Vars[1]
	body := req.HTTP.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		req.HTTPError(500, err)
		return
	}
	req.Redirect("/view/"+title)
}

func main() {
	title_re := "([a-zA-Z0-9]+)$"
	handlers := []web.Handler{
		{"^/$", IndexHandler},
		{"^/view/"+title_re, viewHandler},
		{"^/edit/"+title_re, editHandler},
		{"^/save/"+title_re, saveHandler},
	}

	settings := web.Settings{
		Debug: true,
		TemplatePath: "./templates",
	}

	web.Application(":8080", handlers, &settings)
}
