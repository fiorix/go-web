// Copyright 2013 %template% authors.  All rights reserved.
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
	Debug   bool     `xml:"debug,attr"`
	HTTP    struct {
		Addr     string `xml:"addr,attr"`
		XHeaders bool   `xml:"xheaders,attr"`
	}
	HTTPS struct {
		Addr    string `xml:"addr,attr"`
		CrtFile string
		KeyFile string
	}
	DocumentRoot string
	MySQL        string
	Redis        string
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
	relativePath(basedir, &cfg.HTTPS.CrtFile)
	relativePath(basedir, &cfg.HTTPS.KeyFile)
	relativePath(basedir, &cfg.DocumentRoot)
	return cfg, nil
}

func relativePath(basedir string, path *string) {
	p := *path
	if p != "" && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}
