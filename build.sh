#! /bin/bash

for i in {amd64,arm,arm64}
do
    env GOOS=linux GOARCH="$i" \
        go build -o request-catcher-linux-"$i" request-catcher.go
done
