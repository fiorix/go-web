// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/fiorix/go-redis/redis"
	"github.com/fiorix/go-web/httpxtra"
	"github.com/gorilla/sessions"
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
	// Custom Handler
	handler := httpxtra.Handler{
		Logger:   logger,
		XHeaders: Config.XHeaders,
	}
	// HTTP Server
	server := http.Server{
		Addr:    Config.Addr,
		Handler: handler,
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
			// Use our listener to support Unix sockets.
			log.Fatal(httpxtra.ListenAndServe(server))
			wg.Done()
		}()
	}
	if Config.SSL.Addr != "" {
		wg.Add(1)
		log.Printf("Starting HTTPS server on %s", Config.SSL.Addr)
		go func() {
			https := server
			https.Addr = Config.SSL.Addr
			// No Unix sockets for HTTPS. duh!
			log.Fatal(https.ListenAndServeTLS(
				Config.SSL.CertFile, Config.SSL.KeyFile))
			wg.Done()
		}()
	}
	wg.Wait()
}

func logger(r *http.Request, created time.Time, status, bytes int) {
	fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))
}
