# %name%

Simple web server template with pre-configured MySQL and Redis.

## Preparing the environment

Prerequisites:

- Git
- rsync
- GNU Make
- [Go](http://golang.org) 1.0.3 or newer

First, you should make a copy of this directory, and prepare the new project:

	cp -r simple bozo
	cd bozo
	./bootstrap.sh

Your project is now called **bozo** and is ready to use.

Make sure the Go compiler is installed and `$GOPATH` is set.

Install dependencies, and compile:

	make deps
	make clean
	make all

Generate a self-signed SSL certificate (optional):

	cd ssl
	make

Start Redis if you plan to use it, and set up MySQL (both optional):

	sudo mysql < assets/files/database.sql

Edit the config file and run the server (check MySQL and Redis settings):

	vi %name%.conf
	./%name%

Install, uninstall. Edit Makefile and set PREFIX to the target directory:

	sudo make install
	sudo make uninstall

Allow non-root process to listen on low ports:

	/sbin/setcap 'cap_net_bind_service=+ep' /opt/bozo/server

Good luck!
