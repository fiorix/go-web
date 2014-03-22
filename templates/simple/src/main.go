// Copyright 2013-2014 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	html "html/template"
	text "text/template"

	"github.com/fiorix/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

const (
	VERSION = "1.0"
	APPNAME = "%name%"
)

var (
	Config *ConfigData

	// Templates
	HTML *html.Template
	TEXT *text.Template

	// DBs
	MySQL *sql.DB
	Redis *redis.Client
)

func main() {
	Configfile := flag.String("c", "%name%.conf", "set config file")
	flag.Usage = func() {
		fmt.Println("Usage: %name% [-c %name%.conf]")
		os.Exit(1)
	}
	flag.Parse()

	var err error
	Config, err = LoadConfig(*Configfile)
	if err != nil {
		log.Fatal(err)
	}

	// Parse templates.
	HTML = html.Must(html.ParseGlob(Config.TemplatesDir + "/*.html"))
	TEXT = text.Must(text.ParseGlob(Config.TemplatesDir + "/*.txt"))

	// Set up databases.
	Redis = redis.New(Config.DB.Redis)
	MySQL, err = sql.Open("mysql", Config.DB.MySQL)
	if err != nil {
		log.Fatal(err)
	}

	// Set GOMAXPROCS and show server info.
	var cpuinfo string
	if n := runtime.NumCPU(); n > 1 {
		runtime.GOMAXPROCS(n)
		cpuinfo = fmt.Sprintf("%d CPUs", n)
	} else {
		cpuinfo = "1 CPU"
	}
	log.Printf("%s v%s (%s)", APPNAME, VERSION, cpuinfo)

	// Add routes, and run HTTP and HTTPS servers.
	RouteHTTP()
	wg := &sync.WaitGroup{}
	if Config.HTTP.Addr != "" {
		wg.Add(1)
		go ListenHTTP()
	}
	if Config.HTTPS.Addr != "" {
		wg.Add(1)
		go ListenHTTPS()
	}
	wg.Wait()
}
