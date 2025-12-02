#!/bin/sh

OUTPUT=`/bin/ls /abc 2>&1`
RESULT_CODE_1=$?
echo ${OUTPUT} | go run cmd/logger/main.go --code=${RESULT_CODE_1} --app=App1 --db=./cronlog-store.db

