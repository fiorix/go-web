// Copyright 2013 %name% authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"net/smtp"
	"time"
)

type emailreq struct {
	to  []string
	msg []byte
}

var sendmailc = make(chan *emailreq)

// Sendmail sends an email to the pre-configured SMTP server.
// It does so by sending an email request to the DeliverEmail goroutine.
func SendMail(msg []byte, to ...string) {
	if len(to) > 0 {
		sendmailc <- &emailreq{to, msg}
	}
}

// DeliverEmail goroutine. Reads email requests from the `sendmailc` channel
// send send over to an SMTP.
//
// This is currently a dummy function that has to be updated with your own
// logic or method of sending email.
func DeliverEmail() {
	var (
		req *emailreq
		err error
	)
	for {
		req = <-sendmailc
		go func() {
			for n := 0; n < 3; n++ {
				if err = smtp.SendMail(
					cfg.Email.Addr,
					smtp.PlainAuth(
						"", // Identity
						cfg.Email.User,
						cfg.Email.Passwd,
						cfg.Email.Host,
					),
					cfg.Email.From,
					req.to,
					req.msg,
				); err != nil {
					log.Println("SendMail failed:", err)
					time.Sleep(120 * time.Second)
					continue
				}
				log.Println("SendMail ok, to:", req.to)
				break
			}
		}()
	}
}
