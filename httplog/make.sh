#!/bin/bash

# go build script
# 2019.09.01 smlee@sk.com

rm -f httplog
set -x
find . -name '*.go' -exec goimports -d -w {} \;
golint -min_confidence=0.1 .
go build -v -o httplog
#go build -v -race -o httplog
#go build -v -msan -o httplog

{ set +x; } 2> /dev/null
if [ -f httplog ] ; then
	echo ""
	ls -al httplog
	md5sum httplog
fi
