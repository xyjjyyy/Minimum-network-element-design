#!/bin/bash
go build -o main main.go
echo "build main!"
./main -layerLogo=LNK -deviceId=0 &
./main -layerLogo=PHY -deviceId=0 &
./main -layerLogo=APP -deviceId=1 &
./main -layerLogo=LNK -deviceId=1 &
./main -layerLogo=PHY -deviceId=1 &
./main -layerLogo=APP -deviceId=0

sleep 2

pkill -15 main
