#!/bin/bash

set -x 
set -e

if [[ $(id -u) -ne 0 ]]; then
  echo "must be root to run this script"
  exit 1;
fi

testdir=`pwd`
topdir="$testdir/../../.."
testsdir="$testdir/../.."
rvndir="$testdir/.."

cd $topdir
make distclean
make cleanbuild

cd $testsdir
./install-roles.sh

cd $rvndir
rvn destroy
rvn build
rvn deploy
rvn pingwait server db c0 c1 c1
rvn configure
rvn status
ansible-playbook -i .rvn/ansible-hosts network.yml
ansible-playbook -i .rvn/ansible-hosts testnodes.yml
ansible-playbook -i .rvn/ansible-hosts setup.yml

./test/runtests.sh
