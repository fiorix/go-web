# Copyright 2009-2013 The server Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PREFIX=/opt/%template%

all: server

deps:
	go get -u github.com/fiorix/go-redis/redis
	go get -u github.com/fiorix/go-web/httpxtra
	go get -u github.com/go-sql-driver/mysql
	go get -u github.com/gorilla/sessions

server:
	make -C assets
	(cd src; go build -v -o ../server)

clean:
	rm -f server

install: server
	mkdir -m 750 -p ${PREFIX}
	install -m 750 server ${PREFIX}/server
	install -m 640 server.conf ${PREFIX}
	mkdir -m 750 -p ${PREFIX}/SSL
	install -m 750 SSL/Makefile ${PREFIX}/SSL
	mkdir -m 750 -p ${PREFIX}/assets
	rsync -rupE assets ${PREFIX}
	find ${PREFIX}/assets -type f -exec chmod 640 {} \;
	find ${PREFIX}/assets -type d -exec chmod 750 {} \;
	#chown -R www-data: ${PREFIX}

uninstall:
	rm -rf ${PREFIX}
