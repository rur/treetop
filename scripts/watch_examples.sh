#!/bin/sh

# NOTE: To make this work on Mac I installed `brew instal util-linux`
#       for the process session ID control.

sigint_handler()
{
  kill -- -$PID
}

trap sigint_handler 0

while true; do
  set -x
  setsid go run ./examples &
  PID=$!
  fswatch -1 examples/**/*.go
  # the negative PID is necessary to kill subprocesses
  # see https://unix.stackexchange.com/questions/124127/kill-all-descendant-processes
  kill -- -$PID
  # some throttling is good to allow potentilly multiple files be edited
  sleep 5
done