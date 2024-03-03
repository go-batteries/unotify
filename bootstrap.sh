#!/bin/sh
#
# Do other things
# and run server

mkdir -p /var/log

echo "Start server"
/opt/app/server > /var/log/server.log 2>&1 &

echo "Start Worker"
/opt/app/worker \
  -hcl-dir /opt/app/config/statemachines > /var/log/worker.log 2>&1 &

wait
