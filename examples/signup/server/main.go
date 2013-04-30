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
	"sync"
	"time"

	"github.com/fiorix/go-redis/redis"
	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/sessions"
	_ "github.com/ziutek/mymysql/godrv"
)

const VERSION = "1.0"
const APPNAME = "Foobar"

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
	http.HandleFunc("/signup/", https(unauthenticated(SignUpHandler)))
	http.HandleFunc("/signup/confirm/", SignUpConfirmHandler)
	// Sign In and Out
	http.HandleFunc("/signin/", https(unauthenticated(SignInHandler)))
	http.HandleFunc("/signout/", SignOutHandler)
	// Lost password
	http.HandleFunc("/recovery/", https(unauthenticated(RecoveryHandler)))
	http.HandleFunc("/recovery/confirm/", RecoveryConfirmHandler)
	// Signed In handlers
	http.HandleFunc("/main/", authenticated(MainHandler))
	http.HandleFunc("/settings/", https(authenticated(SettingsHandler)))
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
	log.Printf("%s v%s (%d %s)", APPNAME, VERSION, numCPU, label)
	wg := &sync.WaitGroup{}
	if Config.Addr != "" {
		wg.Add(1)
		log.Printf("Starting HTTP server on %s", Config.Addr)
		go func() {
			log.Fatal(server.ListenAndServe())
			wg.Done()
		}()
	}
	if Config.SSL.Addr != "" {
		wg.Add(1)
		log.Printf("Starting HTTPS server on %s", Config.SSL.Addr)
		go func() {
			https := server
			https.Addr = Config.SSL.Addr
			log.Fatal(https.ListenAndServeTLS(
				Config.SSL.CertFile, Config.SSL.KeyFile))
			wg.Done()
		}()
	}
	wg.Wait()
}

func logger(w http.ResponseWriter, r *http.Request) {
	var s string
	if r.TLS != nil {
		s = "S" // soz no ternary :/
	}
	log.Printf("HTTP%s %d %s %s (%s) :: %s",
		s,
		w.Status(),
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		time.Since(r.Created))
}
