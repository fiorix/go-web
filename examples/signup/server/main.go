// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/fiorix/go-redis/redis"
	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/sessions"
	_ "github.com/ziutek/mymysql/godrv"
)

const VERSION = "1.0"

var Redis *redis.Client
var MySQL *sql.DB
var Config *ConfigData
var Tmpl *Templates
var Session sessions.Store

func main() {
	var err error
	cfgfile := flag.String("config", "config.xml", "set config file")
	sessKey := flag.Bool("keygen", false, "dump random key and exit")
	flag.Parse()
	if *sessKey {
		fmt.Println(RandHex(24))
		return
	}
	Config, err = ReadConfig(*cfgfile)
	if err != nil {
		log.Fatal(err)
	}
	// Load templates
	Tmpl, err = LoadTemplates(Config.TemplatePath, "_base.html")
	if err != nil {
		log.Fatal(err)
	}
	// Set up databases
	Redis = redis.New(Config.Redis)
	MySQL, err = sql.Open("mymysql", Config.MySQL)
	if err != nil {
		log.Fatal(err)
	}
	// Set up session keys
	Session = sessions.NewCookieStore(Config.SessionKey)
	// Public handlers
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/static/", StaticHandler)
	http.HandleFunc("/legal.txt", StaticHandler)
	http.HandleFunc("/favicon.ico", StaticHandler)
	// Sign Up
	http.HandleFunc("/signup/", unauthenticated(SignUpHandler))
	http.HandleFunc("/signup/confirm/", SignUpConfirmHandler)
	// Sign In and Out
	http.HandleFunc("/signin/", unauthenticated(SignInHandler))
	http.HandleFunc("/signout/", SignOutHandler)
	// Lost password
	http.HandleFunc("/recovery/", unauthenticated(RecoveryHandler))
	http.HandleFunc("/recovery/confirm/", RecoveryConfirmHandler)
	// Signed In handlers
	http.HandleFunc("/main/", authenticated(MainHandler))
	http.HandleFunc("/settings/", authenticated(SettingsHandler))
	// HTTP Server
	server := http.Server{
		Addr:     Config.Addr,
		Logger:   logger,
		XHeaders: Config.XHeaders,
	}
	numCPU := runtime.NumCPU()
	label := "CPU"
	if numCPU > 1 {
		label += "s"
	}
	runtime.GOMAXPROCS(numCPU)
	log.Printf("AppServer %s starting on %s (%d %s)",
		VERSION, Config.Addr, numCPU, label)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func logger(w http.ResponseWriter, r *http.Request) {
	log.Printf("HTTP %d %s %s (%s) :: %s",
		w.Status(),
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		time.Since(r.Created))
}
