// Copyright 2013 %name% authors.  All rights reserved.
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

	"github.com/fiorix/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
)

const (
	VERSION = "1.0"
	APPNAME = "%name%"
)

var (
	cfg   *ConfigData
	MySQL *sql.DB
	Redis *redis.Client
)

func main() {
	cfgfile := flag.String("c", "%name%.conf", "set config file")
	flag.Usage = func() {
		fmt.Println("Usage: %name% [-c %name%.conf]")
		os.Exit(1)
	}
	flag.Parse()

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
