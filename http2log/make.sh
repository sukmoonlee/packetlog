#!/bin/bash

# go build script
# 2020.01.28 smlee@sk.com

rm -f http2log
set -x
#go get
find . -name '*.go' -exec goimports -d -w {} \;
golint -min_confidence=0.1 .
go build -v -o http2log
#go build -v -race -o http2log
#go build -v -msan -o http2log

{ set +x; } 2> /dev/null
if [ -f http2log ] ; then
	echo ""
	ls -al http2log
	md5sum http2log
fi
