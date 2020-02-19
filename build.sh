#!/bin/bash -xe

rm -rf dist/
mkdir -p dist/

function build() {
    OS=$1
    ARCH=$2

    mkdir dist/$OS-$ARCH

    cp files/config.yml dist/$OS-$ARCH/config.yml
    cp files/gateway.service dist/$OS-$ARCH/gateway.service

    GOOS=$OS GOARCH=$ARCH go build -o dist/$OS-$ARCH/tellus-market-sdk-gateway main.go
    zip -D -j dist/tellus-market-sdk-gateway-$OS-$ARCH.zip dist/$OS-$ARCH/*
}

build linux amd64
