// Copyright 2013 %name% authors.  All rights reserved.
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
	InviteOnly   bool   `toml:"invite_only"`

	Session struct {
		AuthKey  string `toml:"auth_key"`
		CryptKey string `toml:"crypt_key"`
	} `toml:"session"`

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

	Email struct {
		Addr    string `toml:"smtp_addr"`
		From    string `toml:"from"`
		ReplyTo string `toml:"reply_to"`
		User    string `toml:"plainauth_user"`
		Passwd  string `toml:"plainauth_passwd"`
		Host    string `toml:"plainauth_host"`
	} `toml:"email_client"`
}

// LoadConfig reads and parses the configuration file.
func LoadConfig(filename string) (*ConfigData, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &ConfigData{}
	if _, err := toml.Decode(string(buf), c); err != nil {
		return nil, err
	}

	// Make files' path relative to the config file's directory.
	basedir := filepath.Dir(filename)
	relativePath(basedir, &c.DocumentRoot)
	relativePath(basedir, &c.TemplatesDir)
	relativePath(basedir, &c.HTTPS.CertFile)
	relativePath(basedir, &c.HTTPS.KeyFile)

	return c, nil
}

func relativePath(basedir string, path *string) {
	p := *path
	if p != "" && p[0] != '/' {
		*path = filepath.Join(basedir, p)
	}
}
