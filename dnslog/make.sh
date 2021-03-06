#!/bin/bash

# go build script
# 2019.08.25 smlee@sk.com

rm -f dnslog
set -x

# go get -u golang.org/x/tools/cmd/goimports
find . -name '*.go' -exec goimports -d -w {} \;

# go get -u golang.org/x/lint/golint
golint -min_confidence=0.1 .

go build -v -o dnslog
#go build -v -race -o dnslog
#go build -v -msan -o dnslog

{ set +x; } 2> /dev/null
if [ -f dnslog ] ; then
	echo ""
	ls -al dnslog
	md5sum dnslog
fi
