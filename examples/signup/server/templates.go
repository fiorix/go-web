// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

// Wrapper for html and text templates.

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"os"
	"path/filepath"
	text_template "text/template"

	"github.com/fiorix/go-web/http"
)

// LoadTemplates pre-load all html files under dir, recursively.
// Each directory is expected to have a basefile (usually _base.html) to
// serve as the base container for rendering templates.
func LoadTemplates(dir string, basefile string) (*Templates, error) {
	t := &Templates{BaseDir: dir, BaseFile: basefile}
	return t, t.Scan()
}

// Templates stores all pre-loaded templates.
type Templates struct {
	BaseDir    string
	BaseFile   string
	html_cache map[string]*template.Template
	text_cache map[string]*text_template.Template
}

// Render renders a pre-loaded text or html template.
func (t *Templates) Render(w io.Writer, name string, data interface{}) error {
	html_t, ok := t.html_cache[name]
	if ok {
		return html_t.ExecuteTemplate(w, t.BaseFile, data)
	}
	text_t, ok := t.text_cache[name]
	if ok {
		return text_t.ExecuteTemplate(w, filepath.Base(name), data)
	}
	return errors.New("Template not found: " + name)

}

// Scan recursively parses *.html on each directory under BaseDir.
func (t *Templates) Scan() error {
	t.html_cache = make(map[string]*template.Template)
	t.text_cache = make(map[string]*text_template.Template)
	filepath.Walk(t.BaseDir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			err = t.parseTemplates(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

func (t *Templates) parseTemplates(dir string) error {
	// HTML
	// TODO: reload files on demand
	//fmt.Println("Scanning", dir)
	files, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return err
	}
	for _, name := range files {
		if filepath.Base(name) == t.BaseFile {
			continue
		}
		b := filepath.Join(dir, t.BaseFile)
		k := name[len(t.BaseDir)+1:]
		//fmt.Println("Adding", k)
		// TODO: add "eq" and "or" functions
		v := template.Must(template.ParseFiles(b, name))
		t.html_cache[k] = v
	}
	// Plain text
	// TODO: reload files on demand
	files, err = filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil {
		return err
	}
	for _, name := range files {
		k := name[len(t.BaseDir)+1:]
		//fmt.Println("Adding", k)
		v := text_template.Must(text_template.ParseFiles(name))
		t.text_cache[k] = v
	}
	return nil
}

// renderTemplate renders a template or returns http 500 on failure.
func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if err := Tmpl.Render(w, name, data); err != nil {
		httpError(w, 500, err)
	}
}

// renderTemplateBytes renders a template and returns its bytes.
func renderTemplateBytes(name string, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := Tmpl.Render(&buf, name, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
