Modified version of http package
================================

This is an experimental, slightly modified (for the good) version of Go's
[net/http](http://golang.org/pkg/net/http/) package.

Following is a summary of the new features.

http.Server
-----------

Server can listen on Unix Sockets.
This is useful when the server is proxied by a frontend server like Nginx.

The main reason this feature was added to the Server is to support on service
migration from existing set ups with Nginx.

Example:

	server := http.Server{Addr: "./test.sock"}

or:

	http.ListenAndServe("./test.sock", nil)

Server can overwrite Request.RemoteAddr with the contents of X-Forwarded-For
HTTP header, by setting XHeaders to true.
Useful for servers proxied via Unix Socket or servers sitting behind load
balancers:

	server := http.Server{Addr: ":8080", XHeaders: true}

Server can call a Logger function at the end of every request, by setting
the Logger field while creating the server, or later
(must be an http.HandlerFunc).

Example:

	func logger(w http.ResposeWriter, req *http.Request) {
		log.Printf("HTTP %d %s %s (%s) :: %s",
			w.Status(),
			req.Method,
			req.URL.Path,
			req.RemoteAddr,
			time.Since(req.Created))
	}

	...

	server := http.Server{Addr: ":8080", Logger: logger}
	server.ListenAndServe()

http.ResponseWriter
-------------------

ResponseWriter interface can return the status code of the response by calling
the *Status* method, after it is written (after a call to Write/WriteHeader).

Useful for analytics and logging purposes. See above.

http.Request
------------

Request now have a *Created* field, that represents the time of the creation
of the request. Useful for analytics and logging purposes.

Request also have a new field, *Vars*, to support regexp-based multiplexers.
Request.Vars holds the result of the regexp pattern executed on URL.Path.

See [remux/](remux/remux.go) for details.

Server-Sent events
------------------

The [sse/](sse/sse.go) package provides a simple interface for Server-Sent events,
also known as push notifications.

Regexp-based request multiplexer
--------------------------------

The [remux/](remux/remux.go) package provides a regexp-based multiplexer to route requests
to handlers.

Example:

	remux.HandleFunc("^/(foo|bar)?$", FoobarHandler)
	http.ListenAndServe(":8080", remux.DefaultServeMux)

Handlers can access the capturing groups (e.g.: foo, bar) via *req.Vars*.

Examples
--------

Check out the [examples](https://github.com/fiorix/go-web/tree/master/examples).
