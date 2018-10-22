#!/bin/bash

docker build -t nex-builder -f builder.dock --no-cache .
docker run -v "$topdir/build:/build" nex-builder
