/* Copyright 2013 Alexandre Fiori
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

package main

import "github.com/fiorix/web"

func IndexHandler(req web.RequestHandler) {
	req.Write("Hello, world")
}

func main() {
	handlers := []web.Handler{
		{"/", IndexHandler},
	}
	web.Application(":8080", handlers, &web.Settings{Debug:true})
}
