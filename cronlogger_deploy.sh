#!/bin/sh

TSTAMP=`date -u +%Y%m%d.%H%M%S`

make clean build

cp ./dist/linux/arm64/* ../cronlogger.deployment
cp ./application.yaml ../cronlogger.deployment
cd ../cronlogger.deployment
git add cronlogger
git add cronlogger_server
git add application.yaml
git commit -m "new deployment version ${TSTAMP}"
git push
