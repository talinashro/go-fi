#!/usr/bin/env bash
set -euo pipefail

# 1) Build & launch the orchestrator
(cd examples/orchestrator && go build -o ../../orchestrator .)
./orchestrator --ctrl-addr 0.0.0.0:8081 &
ORCH_PID=$!
sleep 1  # give it a moment to start

# 2) Show initial status
echo
echo "=== Initial status ==="
curl -s http://localhost:8081/status | jq

# 3) Instruct it to fail the first EC2 call
echo
echo "=== Injecting 1 EC2 failure ==="
curl -s "http://localhost:8081/set?key=EC2&count=1"
curl -s http://localhost:8081/status | jq

# 4) Run the workflow (which will now fail once, then succeed)
echo
echo "=== Running workflow ==="
./orchestrator

# 5) Check status again
echo
echo "=== Final status ==="
curl -s http://localhost:8081/status | jq

# 6) Cleanup
kill $ORCH_PID
