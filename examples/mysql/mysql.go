// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

// MySQL demo uses Michał Derkacz's (github.com/ziutek/mymysql) driver.
// It is thread-safe therefore safe to use by multiple requests (goroutines),
// and (re)connects to MySQL on demand. It's also a connection pool.
// See http://golang.org/pkg/database/sql/#DB for detailed information.

import (
	"database/sql"
	"fmt"
	"github.com/fiorix/go-web/http"
	_ "github.com/ziutek/mymysql/godrv"
	_ "github.com/ziutek/mymysql/thrsafe"
	"log"
	"time"
)

var DB *sql.DB

type QueryResult struct {
	one    int
	foobar string
	now    time.Time
}

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	rows, err := DB.Query(`select 1, "foobar", now()`)
	if err != nil {
		// Returns 503 (Service Unavailable) and logs the error
		log.Println("DB.Query failed:", err.Error())
		http.Error(w, http.StatusText(503), 503)
		return
	}
	var qr QueryResult
	for rows.Next() {
		rows.Scan(&qr.one, &qr.foobar, &qr.now)
		fmt.Fprintf(w, "one:%d\r\nfoobar:%s\r\nnow:%s\r\n",
			qr.one, qr.foobar, qr.now)
	}
}

func main() {
	var err error
	dbinfo := "dbname/user/passwd"
	// dbinfo := "tcp:127.0.0.1:3306*dbname/user/passwd" // default tcp
	// dbinfo := "unix:/tmp/mysql.sock*dbname/user/passwd" // performs better
	if DB, err = sql.Open("mymysql", dbinfo); err != nil {
		panic(err)
	}

	// HTTP Server
	http.HandleFunc("/", IndexHandler)
	http.ListenAndServe(":8080", nil)
}
