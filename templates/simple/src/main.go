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
	"os/signal"
	"runtime"
	"syscall"

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
	configFile := flag.String("c", "%name%.conf", "")
	logFile := flag.String("logfile", "", "")
	flag.Usage = func() {
		fmt.Println("Usage: %name% [-c %name%.conf] [-logfile FILE]")
		os.Exit(1)
	}
	flag.Parse()

	var err error
	Config, err = LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize log
	if *logFile != "" {
		setLog(*logFile)
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
	routeHTTP()
	if Config.HTTP.Addr != "" {
		go listenHTTP()
	}
	if Config.HTTPS.Addr != "" {
		go listenHTTPS()
	}
	select {}
}

func setLog(filename string) {
	f := openLog(filename)
	log.SetOutput(f)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP)
	go func() {
		// Recycle log file on SIGHUP.
		var fb *os.File
		for {
			<-sigc
			fb = f
			f = openLog(filename)
			log.SetOutput(f)
			fb.Close()
		}
	}()
}

func openLog(filename string) *os.File {
	f, err := os.OpenFile(
		filename,
		os.O_WRONLY|os.O_CREATE|os.O_APPEND,
		0666,
	)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(err)
	}
	return f
}
