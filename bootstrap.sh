#!/bin/sh
#
# Do other things
# and run server

echo "Start server"
/opt/app/server 2>&1 &

echo "Start Worker"
/opt/app/worker \
  -hcl-dir /opt/app/config/statemachines 2>&1 &

wait
