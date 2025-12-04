#!/bin/bash
set -e

LITESTREAM_VERSION=litestream-0.5.2-linux-x86_64.tar.gz


if [ ! -e "./litestream" ]; then
  # get litestream binary
  curl -L https://github.com/benbjohnson/litestream/releases/download/${LITESTREAM_VERSION} -o litestream.tar.gz
  tar xf litestream.tar.gz && rm litestream.tar.gz
fi

if [ -e "./litestream.tar.gz" ]; then
  rm ./litestream.tar.gz
fi

. ./.env

if [[ -z "${LITESTREAM_ACCESS_KEY_ID}" ]]; then
  echo "LITESTREAM_ACCESS_KEY_ID is required"
  exit 1
fi

if [[ -z "${LITESTREAM_SECRET_ACCESS_KEY}" ]]; then
  echo "LITESTREAM_SECRET_ACCESS_KEY is required"
  exit 1
fi

if [[ -z "${REPLICA_URL}" ]]; then
  echo "REPLICA_URL is required"
  exit 1
fi

echo "restore from >> ${REPLICA_URL}"

if [ -e ../cronlog-store.db ]; then
    echo ".. [CLEANUP] remove the current database file"
	  rm ../cronlog-store.db
fi 

./litestream restore -o ../cronlog-store.db $REPLICA_URL
