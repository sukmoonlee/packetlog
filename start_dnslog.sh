#!/bin/bash

# dnslog start script
# 2019.08.25 smlee@sk.com

################################################################################
# Default Configure
DNSLOG_DEVICE=eth0
DNSLOG_SNAPLEN=1560
DNSLOG_BUFSIZE=32
DNSLOG_FILTER="port 53"
DNSLOG_DIR="/log/dnslog"
DNSLOG_CPUNO=

################################################################################
# Read Configuration File

if [ -f /etc/sysconfig/dnslog ] ; then
	. /etc/sysconfig/dnslog
elif [ -f ./dnslog.conf ] ; then
	. ./dnslog.conf
fi

export DNSLOG_DEVICE
export DNSLOG_SNAPLEN
export DNSLOG_BUFSIZE
export DNSLOG_FILTER
export DNSLOG_DIR
export DNSLOG_CPUNO

case "$1" in
start)
	PID=$(pgrep -x -u root dnslog)
	if [ "$PID" != "" ] ; then
		echo "Another Running Program (PID: $PID)"
		exit 4
	fi
	./dnslog &
	#./dnslog -log_split &
	#./dnslog -cpuprofile dnslog.prof -c 1000000 &
	PID=$!
	rc=$?
	echo "Start Program (PID: $PID)"
	;;
stop)
	PID=$(pgrep -x -u root dnslog)
	if [ "$PID" != "" ] ; then
		echo "Stop Program (PID: $PID)"
		kill -9 "$PID"
		rc=$?
	else
		echo "Not found Running Program"
		exit 2
	fi
	;;
status)
	PID=$(pgrep -x -u root dnslog)
	if [ "$PID" != "" ] ; then
		echo "Status Prorgam (PID: $PID)"
		echo "================================================================================"
		lsof -np "$PID"
		echo "================================================================================"
		ps -fp "$PID"
	else
		echo "Not search Running Program"
		exit 3
	fi
	;;
restart|force-reload)
	$0 stop
	sleep 1
	$0 start
	rc=$?
	;;
*)
	echo $"Usage: $0 {start|stop|status|restart|force-reload}"
	exit 1
esac

exit $rc
