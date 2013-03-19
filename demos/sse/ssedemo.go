// Copyright 2013 Alexandre Fiori
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// This demo is live on http://cos.pe

package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"github.com/nuswit/go-web"
	"html"
	"log"
	"os"
	"strconv"
	"time"
)

type Message struct {
	FrameNo int
	FrameBuf string
}

type Frame struct {
	Time time.Duration
	Buf string  // This is a JSON-encoded Message{FrameBuf:...}
}
var frames []Frame
func loadMovie(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	gzfile, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	lineno := 1
	reader := bufio.NewReader(gzfile)
	frameNo := 1
	frameBuf := ""
	var frameTime time.Duration
	var part string
	for {
		if part, err = reader.ReadString('\n'); err != nil {
			break
		}

		switch(lineno % 14) {
		case 0:
			b := html.EscapeString(frameBuf+part)
			j, _ := json.Marshal(Message{frameNo, b})
			frames = append(frames, Frame{frameTime, string(j)})
			frameNo++
			frameBuf = ""
		case 1:
			s := string(part)
			n, e := strconv.Atoi(s[:len(s)-1])
			if e == nil {
				frameTime = time.Duration(n)*time.Second/10
			}
		default:
			frameBuf += part
		}
		lineno += 1
	}
	return nil
}

func IndexHandler(req *web.RequestHandler) {
	req.ServeFile("./index.html")
}

func SSEHandler(req *web.RequestHandler) {
	sf := 0
	startFrame := req.HTTP.FormValue("startFrame")
	if startFrame != "" {
		sf, _ = strconv.Atoi(startFrame)
	}
	if sf < 0 || sf >= cap(frames) {
		sf = 0
	}
	conn, bufrw, err := req.ServeEvents()
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	// Play the movie, frame by frame
	for n, f := range frames[sf:] {
		req.SendEvent(bufrw, &web.MessageEvent{
					Id:strconv.Itoa(n+1), Data:f.Buf})
		time.Sleep(f.Time)
	}
}

func main() {
	err := loadMovie("./ASCIImation.txt.gz")
	if err != nil {
		log.Println(err)
		return
	}
	handlers := []web.Handler{
		{"^/$", IndexHandler},
		{"^/sse$", SSEHandler},
	}
	settings := web.Settings{
		Debug: true,
		XHeaders: true,
		ReadTimeout: 30*time.Second,
		WriteTimeout: 30*time.Second,
	}
	web.Application(":8080", handlers, &settings)
}
