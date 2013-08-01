#!/bin/bash

name=`basename $PWD`
for f in `find . -type f | grep -vE '(git/|lib/)'`; do
	if [ "`file --mime-type $f | grep -E '(text|xml)'`" != "" ]; then
		sed -e s,%simple%,${name},g < $f > $$.tmp
		mv $$.tmp $f
	fi
done
rm -f bootstrap.sh
