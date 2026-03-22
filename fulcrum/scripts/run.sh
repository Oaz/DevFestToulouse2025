#!/bin/bash

API='http://localhost:8086'
export ADMIN_PASSWORD='whatever'

function waitKey() {
  echo
  read -p "$1"
}

clear
cd ..
go build -o output/devfest2025
./output/devfest2025 &
FULCRUM_PID=$!
sleep 2
echo
for phase_id in {1..18}; do
  waitKey "Press Enter to move to phase $phase_id"
  curl -X 'POST' "$API/forward" \
         -H 'accept: application/json' \
         -H 'Content-Type: application/json' \
         -d "{\"admin_password\": \"$ADMIN_PASSWORD\", \"phase_id\": $phase_id}"
done
waitKey "Press Enter to stop the server"
kill $FULCRUM_PID
