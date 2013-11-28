// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"

	"github.com/fiorix/go-redis/redis"
	"github.com/gorilla/sessions"

	_ "github.com/go-sql-driver/mysql"
)

const (
	VERSION = "1.0"
	APPNAME = "%name%"
)

var (
	cfg          *ConfigData
	MySQL        *sql.DB
	Redis        *redis.Client
	Session      sessions.Store
	DocumentRoot http.Handler
)

func main() {
	cfgfile := flag.String("c", "%name%.conf", "")
	keygen := flag.Bool("k", false, "")
	flag.Usage = func() {
		fmt.Println("Usage: %name% [-k] [-c %name%.conf]")
		os.Exit(1)
	}

	flag.Parse()

	if *keygen {
		fmt.Println(RandHex(16))
		return
	}

	var err error
	cfg, err = LoadConfig(*cfgfile)
	if err != nil {
		log.Fatal(err)
	}

	// Set up databases.
	Redis = redis.New(cfg.DB.Redis)
	MySQL, err = sql.Open("mysql", cfg.DB.MySQL)
	if err != nil {
		log.Fatal(err)
	}

	// Load HTML and plain text templates.
	LoadTemplates()

	// Set up session keys
	Session = sessions.NewCookieStore(
		[]byte(cfg.Session.AuthKey),
		[]byte(cfg.Session.CryptKey),
	)

	// Public html.
	DocumentRoot = http.FileServer(http.Dir(cfg.DocumentRoot))

	// Set GOMAXPROCS and show server info.
	var cpuinfo string
	if n := runtime.NumCPU(); n > 1 {
		runtime.GOMAXPROCS(n)
		cpuinfo = fmt.Sprintf("%d CPUs", n)
	} else {
		cpuinfo = "1 CPU"
	}

	log.Printf("%s v%s (%s)", APPNAME, VERSION, cpuinfo)

	// Start email delivery goroutine.
	go DeliverEmail()

	// Run HTTP and HTTPS servers.
	wg := &sync.WaitGroup{}
	if cfg.HTTP.Addr != "" {
		wg.Add(1)
		go ListenHTTP()
	}
	if cfg.HTTPS.Addr != "" {
		wg.Add(1)
		go ListenHTTPS()
	}
	wg.Wait()
}
