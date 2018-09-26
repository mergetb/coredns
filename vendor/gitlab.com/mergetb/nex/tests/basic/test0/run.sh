#!/bin/bash

set -x
set -e

if [[ $(id -u) -ne 0 ]]; then
  echo "must be root to run this script"
  exit 1;
fi

testdir=`pwd`
topdir="$testdir/../../.."
rvndir="$testdir/.."

echo "building software"
cd $topdir
docker build -t nex-builder -f builder.dock --no-cache .
docker run -v "$topdir/build:/build" nex-builder

echo "deploying and configuring raven topology"
cd $rvndir
rvn destroy
rvn build
rvn deploy
rvn pingwait c0 c1 s0 s1 db sw
rvn configure

./test0/runtests.sh

rvn destroy
rm -rf .rvn
