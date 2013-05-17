# angular-signup-app

Web application template written in Go, with AngularJS and Twitter Bootstrap.

It requires MySQL and Redis for storage, and an SMTP server for sending emails
for sign up, and account recovery.

## Prerequisites

For building the application, it is required to have the following software:

- Git
- rsync
- GNU Make
- [Go](http://golang.org) 1.0.3 or newer
- [Closure-compiler](https://developers.google.com/closure/compiler/)
- [YUIcompressor](http://yui.github.io/yuicompressor/)

## Preparing the environment

A few steps are required before you can use it.
Rename the application and all of its templates:

	./bootstrap.sh bozo

It's now called **bozo**, and you're ready to continue the following steps.

Make sure the Go compiler is installed and ``$GOPATH`` is set.

1. Install dependencies

	make deps

2. Compile the server, minify and compress JS and CSS

	make clean all

3. Generate a self-signed SSL certificate (optional)

	cd SSL
	make

4. Set up MySQL (make sure Redis is running too)

	sudo mysql < assets/files/database.sql

5. Edit the config and run the dev server (check MySQL and Redis settings)

	vi config.xml
	./server/server

6. Install, uninstall. Edit Makefile and set PREFIX to the target directory.

	sudo make install
	sudo make uninstall

7. Allow non-root process to listen on low ports

	/sbin/setcap 'cap_net_bind_service=+ep' /opt/bozo/server
