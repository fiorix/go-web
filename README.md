# go-web

go-web is a collection of utilities for http servers written in
[Go](http://golang.org).

It has the following packages:

### autogzip

- http.Handler that supports on-the-fly gzip encoding
- dummy http client that supports automatic gzip decoding

### httpxtra

- Servers can listen on both TCP or Unix sockets
- Essential request logging (including Apache Common format)
- Support for X-Real-IP and X-Forwarded-For headers for servers sitting behind proxies or load balancers

### remux

- A very simple request multiplexer that supports regular expressions

### sse

- Server-Sent Events library (for push notifications)


## Examples and application templates

Check out the
[examples](https://github.com/fiorix/go-web/tree/master/examples) directory.

There are application templates in the
[templates](https://github.com/fiorix/go-web/tree/master/templates) directory
that can be used as a starting point for new projects. Check them out too.


## Resources

[freegeoip.net](http://freegeoip.net) is a public API for IP geolocation that
uses go-web, and is open source too.

There's a live version of the SSE demo here: <http://cos.pe>
