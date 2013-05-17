// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/sessions"
)

type SearchRequest struct {
	Query string
}

type SearchResultSet struct {
	Phone   sql.NullString
	BrId    sql.NullString
	Name    sql.NullString
	Address sql.NullString
	AddrNo  sql.NullString
	AddrCo  sql.NullString
	NHood   sql.NullString
	ZIP     sql.NullString
	City    sql.NullString
	State   sql.NullString
	RecType sql.NullString
}

type SearchResult struct {
	Phone   string
	BrId    string
	Name    string
	Address string
	AddrNo  string
	AddrCo  string
	NHood   string
	ZIP     string
	City    string
	State   string
	RecType string
}

type SearchResponse struct {
	Ok      bool
	Error   string
	Results []SearchResult
}

func SearchHandler(w http.ResponseWriter, r *http.Request, s *sessions.Session) {
	var v SearchRequest
	err := ParseJSON(r, &v)
	if err != nil {
		httpError(w, 400, err)
		return
	}
	query, args := parse(v.Query)
	stmt, err := MySQL.Prepare(query)
	if err != nil {
		httpError(w, 503, err)
		return
	}
	defer stmt.Close()
	q, err := stmt.Query(args)
	if err != nil {
		httpError(w, 503, err)
		return
	}
	var sr SearchResultSet
	var rows []SearchResult
	for q.Next() {
		err = q.Scan(
			&sr.Phone,
			&sr.BrId,
			&sr.Name,
			&sr.Address,
			&sr.AddrNo,
			&sr.AddrCo,
			&sr.NHood,
			&sr.ZIP,
			&sr.City,
			&sr.State,
			&sr.RecType,
		)
		if err != nil {
			q.Close()
			httpError(w, 503, err)
			return
		} else {
			item := SearchResult{
				sr.Phone.String,
				sr.BrId.String,
				sr.Name.String,
				sr.Address.String,
				sr.AddrNo.String,
				sr.AddrCo.String,
				sr.NHood.String,
				sr.ZIP.String,
				sr.City.String,
				sr.State.String,
				sr.RecType.String,
			}
			rows = append(rows, item)
		}
	}
	JSON(w, SearchResponse{Ok: true, Results: rows})
}

func parse(q string) (string, interface{}) {
	return `select *
		from Record
		where match (Name,Address) against(?)
		limit 200`, q
}
