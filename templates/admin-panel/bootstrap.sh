#!/bin/bash

name=`basename $PWD`
for f in `find . -type f | grep -vE '(git/|vendor/)'`; do
	if [ "`file --mime-type $f | grep -E '(text|xml)'`" != "" ]; then
		sed -e s,%template%,${name},g < $f > $$.tmp
		mv $$.tmp $f
	fi
done
rm -f bootstrap.sh
