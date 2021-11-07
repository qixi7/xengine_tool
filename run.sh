#!/bin/sh

APP_PID_FILE=${APP}.pid
RunArg="-log.dir=./logs/ -log.console=false -prof.mode=cpu,mem"

printUsage(){
	echo "printUsage"
	echo " (+)start:    start your APP"
	echo " (+)stop:     stop your APP"
}

case "$1" in
    -h|-?|h|help)
		printUsage
		;;
	start)
		if [ $# -lt 1 ]; then
			echo "Invalid number of arg"
			exit 1
	    fi
	    if [ -f ${APP_PID_FILE} ] && pid=`cat ${APP_PID_FILE}` && [ -e /proc/${pid} ]; then
	        echo "${APP} alread started..."
	        exit 2
		fi
	    if [ -e ${APP} ]; then
		    ./${APP} ${RunArg} 1>>./stdout.log 2>&1 &
            echo "start ${APP} ok..."
		else
		    echo "start error: ${APP} not exist"
		    exit 3
		fi
		;;
	stop)
		if [ $# -lt 1 ]; then
			echo "Invalid number of arg"
			exit 1
		fi
	    if [ -f ${APP_PID_FILE} ] && pid=`cat ${APP_PID_FILE}` && [ -e /proc/${pid} ]; then
	        kill ${pid}
            echo "stop ${APP} ok..."
		else
		    echo "stop error: ${APP} alread stopped..."
		    exit 2
		fi
		;;
	*)
		echo "unknown Operation: $1"
		printUsage
		;;
esac

exit 0