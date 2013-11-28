# Templates

go-web ships with some application templates (server skeleton) that serve as a
starting point for new projects.

They usually have:

- Makefile to build the server, and some times other assets (e.g. compress js and css)
- Sample configuration file (for server port, document root, etc)
- HTTP and HTTPS on the same server
- Scripts to create self-signed certificates
- Pre-configured Redis and MySQL

## Using templates

Just make a copy of the template and run its `bootstrap.sh` script.

Example:

	cp -r simple foobar
	cd foobar
	./bootstrap.sh

Install dependencies:

	make deps

Compile and run the server:

	make clean all
	./foobar

Each template has its own README.md with more details.
