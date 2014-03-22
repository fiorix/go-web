# Copyright 2013-2014 %name% Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

NAME=%name%

all: deps server

deps:
	go get -v

server:
	go build -v -o %name%

clean:
	rm -f %name%

.PHONY: server
