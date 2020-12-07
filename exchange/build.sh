#!/usr/bin/env bash

set -x

GitCommitLog=`git log --pretty=oneline -n 1`
GitCommitLog=${GitCommitLog//\'/\"}
GitStatus=`git status -s`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

LDFlags=" \
    -X 'github.com/Qitmeer/exchange-lib/exchange/version.GitCommitLog=${GitCommitLog}' \
    -X 'github.com/Qitmeer/exchange-lib/exchange/version.GitStatus=${GitStatus}' \
    -X 'github.com/Qitmeer/exchange-lib/exchange/version.BuildTime=${BuildTime}' \
    -X 'github.com/Qitmeer/exchange-lib/exchange/version.BuildGoVersion=${BuildGoVersion}' \
"

ROOT_DIR=`pwd`

if [ ! -d ${ROOT_DIR}/bin ]; then
  mkdir bin
fi

cd ${ROOT_DIR} && GOOS=linux GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/linux/exchange &&
cd ${ROOT_DIR} && GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFlags" -o ${ROOT_DIR}/bin/darwin/exchange &&
cp ${ROOT_DIR}/config.toml ${ROOT_DIR}/bin/linux/config.toml &&
cp ${ROOT_DIR}/config.toml ${ROOT_DIR}/bin/darwin/config.toml &&
ls -lrt ${ROOT_DIR}/bin &&
cd ${ROOT_DIR} && ./bin/darwin/exchange -v &&

echo 'build done.'