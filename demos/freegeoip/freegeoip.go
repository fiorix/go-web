// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"database/sql"
	_	"github.com/mattn/go-sqlite3"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/fiorix/web"
	"net"
	"path/filepath"
	"regexp"
	"strings"
)

type GeoIP struct {
    XMLName	xml.Name	`json:"-" xml:"Response"`
    Ip		string		`json:"ip"`
    CountryCode string		`json:"country_code"`
    CountryName string		`json:"country_name"`
    RegionCode	string		`json:"region_code"`
    RegionName	string		`json:"region_name"`
    CityName	string		`json:"city" xml:"City"`
    ZipCode	string		`json:"zipcode"`
    Latitude	float32		`json:"latitude"`
    Longitude	float32		`json:"longitude"`
    MetroCode	string		`json:"metro_code"`
    AreaCode	string		`json:"areacode"`
}

// http://en.wikipedia.org/wiki/Reserved_IP_addresses
var reservedIPs = []net.IPNet{
	{net.IPv4(0, 0, 0, 0),		net.IPv4Mask(255, 0, 0, 0)},
	{net.IPv4(0, 0, 0, 0),		net.IPv4Mask(255, 0, 0, 0)},
	{net.IPv4(10, 0, 0, 0),		net.IPv4Mask(255, 192, 0, 0)},
	{net.IPv4(100, 64, 0, 0),	net.IPv4Mask(255, 0, 0, 0)},
	{net.IPv4(127, 0, 0, 0),	net.IPv4Mask(255, 0, 0, 0)},
	{net.IPv4(169, 254, 0, 0),	net.IPv4Mask(255, 255, 0, 0)},
	{net.IPv4(172, 16, 0, 0),	net.IPv4Mask(255, 240, 0, 0)},
	{net.IPv4(192, 0, 0, 0),	net.IPv4Mask(255, 255, 255, 248)},
	{net.IPv4(192, 0, 2, 0),	net.IPv4Mask(255, 255, 255, 0)},
	{net.IPv4(192, 88, 99, 0),	net.IPv4Mask(255, 255, 255, 0)},
	{net.IPv4(192, 168, 0, 0),	net.IPv4Mask(255, 255, 0, 0)},
	{net.IPv4(198, 18, 0, 0),	net.IPv4Mask(255, 254, 0, 0)},
	{net.IPv4(198, 51, 100, 0),	net.IPv4Mask(255, 255, 255, 0)},
	{net.IPv4(203, 0, 113, 0),	net.IPv4Mask(255, 255, 255, 0)},
	{net.IPv4(224, 0, 0, 0),	net.IPv4Mask(240, 0, 0, 0)},
	{net.IPv4(240, 0, 0, 0),	net.IPv4Mask(240, 0, 0, 0)},
	{net.IPv4(255, 255, 255, 255),	net.IPv4Mask(255, 255, 255, 255)},
}

func LookupHandler(req web.RequestHandler, db *sql.DB) {
        req.SetHeader("Access-Control-Allow-Origin", "*")
	format, addr := req.Vars[1], req.Vars[2]
	if addr == "" {
		addr = strings.Split(req.HTTP.RemoteAddr, ":")[0]
	} else {
		addrs, err := net.LookupHost(addr)
		if err != nil {
			req.HTTPError(400, err)
			return
		}
		addr = addrs[0]
	}

	IP := net.ParseIP(addr)
	geoip := GeoIP{Ip: addr}

	reserved := false
	for _, net := range reservedIPs {
		if net.Contains(IP) {
			reserved = true
			break
		}
	}

	if reserved {
		geoip.CountryCode = "RD"
		geoip.CountryName = "Reserved"
	} else {
		q := "SELECT "+
		"  city_location.country_code, country_blocks.country_name, "+
		"  city_location.region_code, region_names.region_name, "+
		"  city_location.city_name, city_location.postal_code, "+
		"  city_location.latitude, city_location.longitude, "+
		"  city_location.metro_code, city_location.area_code "+
		"FROM city_blocks "+
		"  NATURAL JOIN city_location "+
		"  INNER JOIN country_blocks ON "+
		"    city_location.country_code = country_blocks.country_code "+
		"  INNER JOIN region_names ON "+
		"    city_location.country_code = region_names.country_code "+
		"    AND "+
		"    city_location.region_code = region_names.region_code "+
		"WHERE city_blocks.ip_start <= ? "+
		"ORDER BY city_blocks.ip_start DESC LIMIT 1"

		stmt, err := db.Prepare(q)
		if err != nil {
			req.HTTPError(500, err)
			return
		}

		defer stmt.Close()

		var uintIP uint32
		b := bytes.NewBuffer(IP.To4())
		binary.Read(b, binary.BigEndian, &uintIP)
		err = stmt.QueryRow(uintIP).Scan(
			&geoip.CountryCode,
			&geoip.CountryName,
			&geoip.RegionCode,
			&geoip.RegionName,
			&geoip.CityName,
			&geoip.ZipCode,
			&geoip.Latitude,
			&geoip.Longitude,
			&geoip.MetroCode,
			&geoip.AreaCode)
		if err != nil {
			req.HTTPError(500, err)
			return
		}
	}

	switch format[0] {
	case 'c':
		req.SetHeader("Content-Type", "application/csv")
		req.Write("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\""+
			  "\"%s\",\"%0.4f\",\"%0.4f\",\"%s\",\"%s\"\r\n",
			  geoip.Ip,
			  geoip.CountryCode, geoip.CountryName,
			  geoip.RegionCode, geoip.RegionName,
			  geoip.CityName, geoip.ZipCode,
			  geoip.Latitude, geoip.Longitude,
			  geoip.MetroCode, geoip.AreaCode)
	case 'j':
		req.SetHeader("Content-Type", "application/json")
		resp, err := json.Marshal(geoip)
		if err != nil {
			req.HTTPError(500, err)
			return
		}
		req.Write("%s\r\n", resp)
	case 'x':
		req.SetHeader("Content-Type", "application/xml")
		resp, err := xml.MarshalIndent(geoip, " ", " ")
		if err != nil {
			req.HTTPError(500, err)
			return
		}
		req.Write("<?xml version=\"1.0\" encoding=\"UTF-8\"?>"+
			  "%s\r\n", resp)
	}
}

func makeHandler(db *sql.DB,
		 fn func(web.RequestHandler, *sql.DB)) web.HandlerFunc {
	return func(req web.RequestHandler) { fn(req, db) }
}

// This is just for backwards compatibility with freegeoip.net
func IndexHandler(req web.RequestHandler) {
	req.Redirect("/static/index.html")
}

var static_re = regexp.MustCompile("..[/\\\\]")  // gtfo
func StaticHandler(req web.RequestHandler) {
	filename := req.Vars[1]
	if static_re.MatchString(filename) {
		req.NotFound()
		return
	}
	req.ServeFile(filepath.Join("./static", filename))
}

func main() {
	db, err := sql.Open("sqlite3", "db/ipdb.sqlite")
	if err != nil {
		fmt.Println(err)
		return
	}
	handlers := []web.Handler{
		{"^/$", IndexHandler},
		{"^/static/(.*)$", StaticHandler},
		{"^/(crossdomain.xml)$", StaticHandler},
		{"^/(csv|json|xml)/(.*)$", makeHandler(db, LookupHandler)},
	}
	web.Application(":8080", handlers,
			&web.Settings{Debug:true, XHeaders:false})
}
