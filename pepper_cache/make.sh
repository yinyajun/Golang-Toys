#!/usr/bin/env bash
if [ ! -d "$GOROOT" ]; then
		GOROOT="/usr/local/go"
fi

# 设置临时GOPATH
CURDIR=`pwd`
export GOROOT=$GOROOT
OLDGOPATH=$GOPATH
export GOPATH=$CURDIR
echo $GOPATH

# 编译server
echo "building server"
$GOROOT/bin/go install cache/server
if [ $? -ne "0" ]; then
    echo "install server failed"
    exit -1
fi

# 编译client
echo "building client"
$GOROOT/bin/go install cache/client
if [ $? -ne "0" ]; then
    echo "install client failed"
    exit -1
fi

# 编译benchmark
echo "building benchmark"
$GOROOT/bin/go install cache/benchmark
if [ $? -ne "0" ]; then
    echo "install benchmark failed"
    exit -1
fi