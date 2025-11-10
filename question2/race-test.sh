#!/bin/bash

BASE_URL="http://localhost:8080"
#Tujuan script:
# Testing apakah server bisa handle 50 concurrent reads + 50 concurrent writes tanpa error (race condition).
echo "Testing concurrent reads AND writes..."

# Start 50 concurrent reads
for i in {1..50}; do
  curl -s $BASE_URL/users/1 > /dev/null &
done

# Start 50 concurrent writes (at the same time!)
for i in {1..50}; do
  curl -X POST $BASE_URL/users \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"User$i\",\"email\":\"user$i@example.com\"}" \
    -s > /dev/null &
done

wait
echo "Test completed"
