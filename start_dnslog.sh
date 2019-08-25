#!/bin/bash

# dnslog start script
# 2019.08.25 smlee@sk.com

################################################################################
# Default Configure
DNSLOG_DEVICE=eth0
DNSLOG_SNAPLEN=1560
DNSLOG_BUFSUZE=8
DNSLOG_FILTER="port 53"
DNSLOG_DIR="/tmp"

################################################################################
# Read Configuration File

if [ -f /etc/sysconfig/dnslog ] ; then
	. /etc/sysconfig/dnslog
elif [ -f ./dnslog.conf ] ; then
	. ./dnslog.conf
fi

export DNSLOG_DEVICE
export DNSLOG_SNAPLEN
export DNSLOG_BUFSUZE
export DNSLOG_FILTER
export DNSLOG_DIR

case "$1" in
start)
	#./dnslog &
	./dnslog -log_split &
	rc=$?
	;;
stop)
	PID=$(pgrep -x -u root dnslog)
	if [ "$PID" != "" ] ; then
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
		lsof -np "$PID"
		ps -fp "$PID"
	else
		echo "Not search Running Program"
		exit 3
	fi
	;;
restart|force-reload)
	$0 stop
	$0 start
	rc=$?
	;;
*)
	echo $"Usage: $0 {start|stop|status|restart|force-reload}"
	exit 1
esac

exit $rc
