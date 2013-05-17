// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
)

type ConfigData struct {
	XMLName xml.Name `xml:"Server"`
	Debug   bool

	// http
	Addr     string `xml:",attr"`
	XHeaders bool   `xml:",attr"`

	SSL struct {
		Addr     string `xml:",attr"`
		CertFile string
		KeyFile  string
	}

	// settings
	SessionKey []byte

	// assets
	StaticPath   string
	TemplatePath string

	// databases
	MySQL string
	Redis string

	// smtp
	SMTP struct {
		XMLName   xml.Name
		Addr      string
		From      string
		ReplyTo   string
		PlainAuth struct {
			User   string
			Passwd string
			Host   string
		}
	}
}

// ReadConfig reads and parses the XML configuration file.
func ReadConfig(filename string) (*ConfigData, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &ConfigData{}
	if err := xml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}
	// Make file paths relative to the config file's dir.
	basedir := filepath.Dir(filename)
	relativePath(basedir, &cfg.SSL.CertFile)
	relativePath(basedir, &cfg.SSL.KeyFile)
	relativePath(basedir, &cfg.StaticPath)
	relativePath(basedir, &cfg.TemplatePath)
	return cfg, nil
}

func relativePath(basedir string, path *string) {
	p := *path
	if p != "" && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}
