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
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type QueryResult struct {
	one    int
	foobar string
	now    time.Time
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
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
	dbinfo := "user:passwd@tcp(localhost:3306)/dbname?autocommit=true"
	if DB, err = sql.Open("mysql", dbinfo); err != nil {
		panic(err)
	}
	// HTTP Server
	http.HandleFunc("/", IndexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
