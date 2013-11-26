// Copyright 2013 %template% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type ConfigData struct {
	Debug        bool   `toml:"debug"`
	TemplatesDir string `toml:"templates_dir"`
	DocumentRoot string `toml:"document_root"`

	DB struct {
		MySQL string `toml:"mysql"`
		Redis string `toml:"redis"`
	} `toml:"db"`

	HTTP struct {
		Addr     string `toml:"addr"`
		XHeaders bool   `toml:"xheaders"`
	} `toml:"http_server"`

	HTTPS struct {
		Addr     string `toml:"addr"`
		CertFile string `toml:"cert_file"`
		KeyFile  string `toml:"key_file"`
	} `toml:"https_server"`
}

// LoadConfig reads and parses the configuration file.
func LoadConfig(filename string) (*ConfigData, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &ConfigData{}
	if _, err := toml.Decode(string(buf), cfg); err != nil {
		return nil, err
	}
	// Make file paths relative to the config file's dir.
	basedir := filepath.Dir(filename)
	relativePath(basedir, &cfg.DocumentRoot)
	relativePath(basedir, &cfg.TemplatesDir)
	relativePath(basedir, &cfg.HTTPS.CertFile)
	relativePath(basedir, &cfg.HTTPS.KeyFile)
	return cfg, nil
}

func relativePath(basedir string, path *string) {
	p := *path
	if p != "" && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}
