#!/bin/bash

if [ "x$1" == "x" ]; then
	echo use: $0 name
	exit 1
fi
for f in `find . -type f | grep -vE '(git/|lib/)'`; do
	if [ "`file -I $f | grep -E '(text|xml)'`" != "" ]; then
		sed -e s,angular-signup-app,$1,g < $f > $$.tmp
		mv $$.tmp $f
	fi
done
rm -f bootstrap.sh
