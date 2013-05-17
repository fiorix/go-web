#!/bin/bash

echo Set \"Common Name\" to localhost.

echo -- key
openssl genrsa -des3 -out server.key 1024

echo -- csr
openssl req -new -key server.key -out server.csr

echo -- remove passphrase
cp server.key orig.server.key
openssl rsa -in orig.server.key -out server.key

echo -- generate crt
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt
