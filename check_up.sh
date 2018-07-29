#!/bin/bash

SERVICE_STATUS=`ps aux | grep /root/announcer/announcer | wc -l`

if [ "$SERVICE_STATUS" -lt "2" ]; then
	echo "service is down"
	service announcer start
fi


EJABBERD_CONNECTION=`netstat -nap | grep :9999 | wc -l`

if [ "$EJABBERD_CONNECTION" -lt "1" ]; then
        echo "ejabberd connection lost"
        service announcer restart
fi


