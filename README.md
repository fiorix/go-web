# go-web

go-web is a collection of utilities for http servers written in the
[Go](http://golang.org) programming language.

It has the following packages:

- autogzip: An http.Handler that supports on-the-fly gzip encoding, and a client that supports decoding.
- httpxtra: Servers can listen on unix sockets, essential logging including Apache Common format, and support for X-Real-IP and X-Forwarded-For headers - for servers sitting behind proxies or load balancers.
- remux: A very simple request multiplexer that supports regular expressions.
- sse: Server-Sent Events library (for push notifications).

## Examples and application templates

Check out the [examples](https://github.com/fiorix/go-web/tree/master/examples) directory.

There are application templates in the [templates](https://github.com/fiorix/go-web/tree/master/templates) directory that can be used as starting point for new projects. Check them out too.

## Resources

[freegeoip.net](http://freegeoip.net) is a public API for IP geolocation that
uses go-web, and is open source too.

There's a live version of the SSE demo here: <http://cos.pe>
