#!/bin/bash

if [[ ! -f build/coredns ]]; then
  curl -o build/coredns \
    -L https://github.com/mergetb/coredns/releases/download/v1.2.6-nex/coredns
  chmod +x build/coredns
fi

snapcraft clean
snapcraft
mv *.snap build/
