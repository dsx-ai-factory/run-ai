#!/bin/sh

/bin/promtool-agent &
AGENT_PID=$!

echo spawened promtool_agent with PID: ${AGENT_PID}

cleanup(){
  kill -9 $AGENT_PID
  echo killed $AGENT_PID
  exit
}

trap cleanup EXIT

/bin/prometheus "$@"
