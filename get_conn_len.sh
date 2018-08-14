#!/bin/bash

FILE=/etc/zabbix/connections

echo "ejabberd "$(netstat -nap | grep :9999 | wc -l) > $FILE
echo "http "$(netstat -nap | grep :8080 | wc -l) >> $FILE
echo "redis "$(netstat -nap | grep :6379 | wc -l) >> $FILE
echo "user-activity-api "$(netstat -nap | grep :8888 | wc -l) >> $FILE
echo "mysql "$(netstat -nap | grep :33061 | wc -l) >> $FILE
echo "zookeeper "$(netstat -nap | grep :2181 | wc -l) >> $FILE
echo "kafka "$(netstat -nap | grep :9092 | wc -l) >> $FILE
echo "user-activity-rest-api "$(netstat -nap | grep :8888 | wc -l) >> $FILE
