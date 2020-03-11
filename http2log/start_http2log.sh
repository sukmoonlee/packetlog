#!/bin/bash

# http2log start script
# 2020.03.11 smlee@sk.com

################################################################################
# Default Configure
HTTP2LOG_DEVICE=eth1
HTTP2LOG_SNAPLEN=1560
HTTP2LOG_BUFSIZE=32
#HTTP2LOG_FILTER="port 80 and tcp"
HTTP2LOG_DIR="/tmp"
HTTP2LOG_CPUNO=
HTTP2LOG_PORT=80

################################################################################
# Read Configuration File

if [ -f /etc/sysconfig/http2log ] ; then
	. /etc/sysconfig/http2log
elif [ -f ./http2log.conf ] ; then
	. ./http2log.conf
fi

export HTTP2LOG_DEVICE
export HTTP2LOG_SNAPLEN
export HTTP2LOG_BUFSIZE
#export HTTP2LOG_FILTER
export HTTP2LOG_DIR
export HTTP2LOG_CPUNO
export HTTP2LOG_PORT

case "$1" in
start)
	PID=$(pgrep -x -u root http2log)
	if [ "$PID" != "" ] ; then
		echo "Another Running HTTP2LOG Program (PID: $PID)"
		exit 4
	fi
	./http2log &
	#./http2log -log_split &
	#./http2log -cpuprofile http2log.prof -c 1000000 &
	rc=$?
	PID=$!
	if [ "$rc" == "0" ] ; then
		echo "Start HTTP2LOG Program (PID: $PID)"
	else
		echo "Fail run HTTP2LOG Program (PID: $PID, Exit Code:$rc)"
	fi
	;;
stop)
	PID=$(pgrep -x -u root http2log)
	if [ "$PID" != "" ] ; then
		echo "Stop HTTP2LOG Program (PID: $PID)"
		kill -9 "$PID"
		rc=$?
	else
		echo "Not found Running HTTP2LOG Program"
		exit 2
	fi
	;;
status)
	PID=$(pgrep -x -u root http2log)
	if [ "$PID" != "" ] ; then
		echo "Status HTTP2LOG Prorgam (PID: $PID)"
		echo "================================================================================"
		lsof -np "$PID"
		echo "================================================================================"
		ps -fp "$PID"
	else
		echo "Not search Running HTTP2LOG Program"
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
