c0_ip=`rvn ip c0`
c1_ip=`rvn ip c1`

echo "running tests"
avocado run \
  --remote-username rvn \
  --remote-password rvn \
  --remote-hostname $c0_ip \
  /tmp/nex/tests/basic/test0/c0.py

avocado run \
  --remote-username rvn \
  --remote-password rvn \
  --remote-hostname $c1_ip \
  /tmp/nex/tests/basic/test0/c1.py
