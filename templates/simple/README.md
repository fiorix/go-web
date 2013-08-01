# %simple%

Simple web server template written in Go.

It ships with pre-configured MySQL and Redis support.

## Prerequisites

For building the application, it is required to have the following software:

- Git
- rsync
- GNU Make
- [Go](http://golang.org) 1.0.3 or newer

## Preparing the environment

First, you should make a copy of this directory, and prepare your new
project:

	cp -r admin-template bozo
	cd bozo
	./bootstrap.sh

Your project is now called **bozo**, and you're ready to continue the
following steps.

Make sure the Go compiler is installed and ``$GOPATH`` is set.

Install dependencies:

	make deps

Compile the server:

	make clean all

Generate a self-signed SSL certificate (optional):

	cd SSL
	make

Start Redis if you plan to use it, and set up MySQL (both optional):

	sudo mysql < assets/files/database.sql

Edit the config and run the dev server (check MySQL and Redis settings):

	vi server.conf
	./server

Install, uninstall. Edit Makefile and set PREFIX to the target directory:

	sudo make install
	sudo make uninstall

Allow non-root process to listen on low ports:

	/sbin/setcap 'cap_net_bind_service=+ep' /opt/bozo/server

Good luck!
