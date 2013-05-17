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
	"text/template"
	"time"

	"github.com/fiorix/go-redis/redis"
	"github.com/fiorix/go-web/httpxtra"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

const (
	VERSION = "1.0"
	APPNAME = "Foobar"
)

var (
	Config  *ConfigData
	MySQL   *sql.DB
	Redis   *redis.Client
	Tmpl    *template.Template
	Session sessions.Store
)

func route() {
	// Public handlers
	http.Handle("/", http.FileServer(http.Dir(Config.DocumentRoot)))
	http.HandleFunc("/signin.json", SigninHandler)
	http.HandleFunc("/signout/", SignoutHandler)
	http.HandleFunc("/signup.json", SignupHandler)
	http.HandleFunc("/signup-confirm.json", SignupConfirmHandler)
	http.HandleFunc("/recovery.json", RecoveryHandler)
	http.HandleFunc("/recovery-confirm.json", RecoveryConfirmHandler)

	// Private handlers (only for authenticated users)
	http.Handle("/u/", http.StripPrefix("/u/",
		authenticated(UserFS(Config.UsersDocumentRoot))))
	http.HandleFunc("/u/index.json", authenticated(UserIndexHandler))
	http.HandleFunc("/u/search.json", authenticated(SearchHandler))
	http.HandleFunc("/u/settings.json", authenticated(UserSettingsHandler))
}

func hello() {
	numCPU := runtime.NumCPU()
	label := "CPU"
	if numCPU > 1 {
		label += "s"
	}
	runtime.GOMAXPROCS(numCPU)
	log.Printf("%s v%s (%d %s)", APPNAME, VERSION, numCPU, label)
}

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
	Tmpl = template.Must(template.ParseGlob(Config.TemplatesDirectory))
	if err != nil {
		log.Fatal(err)
	}
	// Set up databases
	Redis = redis.New(Config.Redis)
	MySQL, err = sql.Open("mysql", Config.MySQL)
	if err != nil {
		log.Fatal(err)
	}
	// Set up session keys
	Session = sessions.NewCookieStore(Config.SessionKey)
	// Set up routing and print server info
	route()
	hello()
	// Run HTTP and HTTPS servers
	wg := &sync.WaitGroup{}
	if Config.HTTP.Addr != "" {
		wg.Add(1)
		log.Printf("Starting HTTP server on %s", Config.HTTP.Addr)
		go func() {
			// Use httpxtra's listener to support Unix sockets.
			server := http.Server{
				Addr: Config.HTTP.Addr,
				Handler: httpxtra.Handler{
					Logger:   logger,
					XHeaders: Config.HTTP.XHeaders,
				},
			}
			log.Fatal(httpxtra.ListenAndServe(server))
			wg.Done()
		}()
	}
	if Config.HTTPS.Addr != "" {
		wg.Add(1)
		log.Printf("Starting HTTPS server on %s", Config.HTTPS.Addr)
		go func() {
			server := http.Server{
				Addr:    Config.HTTPS.Addr,
				Handler: httpxtra.Handler{Logger: logger},
			}
			log.Fatal(server.ListenAndServeTLS(
				Config.HTTPS.CrtFile, Config.HTTPS.KeyFile))
			wg.Done()
		}()
	}
	wg.Wait()
}

func logger(r *http.Request, created time.Time, status, bytes int) {
	fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))
}
