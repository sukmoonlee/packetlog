#!/bin/bash

# httplog start script
# 2019.09.01 smlee@sk.com

################################################################################
# Default Configure
HTTPLOG_DEVICE=eth0
HTTPLOG_SNAPLEN=1560
HTTPLOG_BUFSIZE=32
HTTPLOG_FILTER="port 80 and tcp"
HTTPLOG_DIR="/tmp"
HTTPLOG_CPUNO=
HTTPLOG_PORT=80

################################################################################
# Read Configuration File

if [ -f /etc/sysconfig/httplog ] ; then
	. /etc/sysconfig/httplog
elif [ -f ./httplog.conf ] ; then
	. ./httplog.conf
fi

export HTTPLOG_DEVICE
export HTTPLOG_SNAPLEN
export HTTPLOG_BUFSIZE
export HTTPLOG_FILTER
export HTTPLOG_DIR
export HTTPLOG_CPUNO
export HTTPLOG_PORT

case "$1" in
start)
	PID=$(pgrep -x -u root httplog)
	if [ "$PID" != "" ] ; then
		echo "Another Running HTTPLOG Program (PID: $PID)"
		exit 4
	fi
	./httplog &
	#./httplog -log_split &
	#./httplog -cpuprofile httplog.prof -c 1000000 &
	rc=$?
	PID=$!
	if [ "$rc" == "0" ] ; then
		echo "Start HTTPLOG Program (PID: $PID)"
	else
		echo "Fail run HTTPLOG Program (PID: $PID, Exit Code:$rc)"
	fi
	;;
stop)
	PID=$(pgrep -x -u root httplog)
	if [ "$PID" != "" ] ; then
		echo "Stop HTTPLOG Program (PID: $PID)"
		kill -9 "$PID"
		rc=$?
	else
		echo "Not found Running HTTPLOG Program"
		exit 2
	fi
	;;
status)
	PID=$(pgrep -x -u root httplog)
	if [ "$PID" != "" ] ; then
		echo "Status HTTPLOG Prorgam (PID: $PID)"
		echo "================================================================================"
		lsof -np "$PID"
		echo "================================================================================"
		ps -fp "$PID"
	else
		echo "Not search Running HTTPLOG Program"
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
