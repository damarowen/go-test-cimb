#!/bin/bash

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Testing User Management REST API${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Test 1: Create a user
echo -e "${BLUE}Test 1: Creating a user${NC}"
RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}')
echo "Response: $RESPONSE"
USER_ID=$(echo $RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
if [ ! -z "$USER_ID" ]; then
  echo -e "${GREEN}✓ User created with ID: $USER_ID${NC}\n"
else
  echo -e "${RED}✗ Failed to create user${NC}\n"
  exit 1
fi

# Test 2: Get the user
echo -e "${BLUE}Test 2: Getting user by ID${NC}"
RESPONSE=$(curl -s -X GET $BASE_URL/users/$USER_ID)
echo "Response: $RESPONSE"
if echo $RESPONSE | grep -q "john@example.com"; then
  echo -e "${GREEN}✓ User retrieved successfully${NC}\n"
else
  echo -e "${RED}✗ Failed to retrieve user${NC}\n"
fi

# Test 3: Update the user
echo -e "${BLUE}Test 3: Updating user${NC}"
RESPONSE=$(curl -s -X PUT $BASE_URL/users/$USER_ID \
  -H "Content-Type: application/json" \
  -d '{"name":"John Smith","email":"john.smith@example.com"}')
echo "Response: $RESPONSE"
if echo $RESPONSE | grep -q "john.smith@example.com"; then
  echo -e "${GREEN}✓ User updated successfully${NC}\n"
else
  echo -e "${RED}✗ Failed to update user${NC}\n"
fi

# Test 4: Create another user
echo -e "${BLUE}Test 4: Creating another user${NC}"
RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com"}')
echo "Response: $RESPONSE"
USER_ID_2=$(echo $RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
if [ ! -z "$USER_ID_2" ]; then
  echo -e "${GREEN}✓ Second user created with ID: $USER_ID_2${NC}\n"
else
  echo -e "${RED}✗ Failed to create second user${NC}\n"
fi

# Test 5: Delete the first user
echo -e "${BLUE}Test 5: Deleting user${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE $BASE_URL/users/$USER_ID)
echo "HTTP Status Code: $HTTP_CODE"
if [ "$HTTP_CODE" = "204" ]; then
  echo -e "${GREEN}✓ User deleted successfully${NC}\n"
else
  echo -e "${RED}✗ Failed to delete user${NC}\n"
fi

# Test 6: Try to get deleted user (should fail)
echo -e "${BLUE}Test 6: Trying to get deleted user (should fail)${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/users/$USER_ID)
echo "HTTP Status Code: $HTTP_CODE"
if [ "$HTTP_CODE" = "404" ]; then
  echo -e "${GREEN}✓ Correctly returns 404 for deleted user${NC}\n"
else
  echo -e "${RED}✗ Should return 404 for deleted user${NC}\n"
fi

# Test 7: Validation - Empty name
echo -e "${BLUE}Test 7: Validation test - Empty name${NC}"
RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"","email":"test@example.com"}')
echo "Response: $RESPONSE"
if echo $RESPONSE | grep -q "error"; then
  echo -e "${GREEN}✓ Correctly validates empty name${NC}\n"
else
  echo -e "${RED}✗ Should reject empty name${NC}\n"
fi

# Test 8: Validation - Invalid email
echo -e "${BLUE}Test 8: Validation test - Invalid email${NC}"
RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"invalid-email"}')
echo "Response: $RESPONSE"
if echo $RESPONSE | grep -q "error"; then
  echo -e "${GREEN}✓ Correctly validates email format${NC}\n"
else
  echo -e "${RED}✗ Should reject invalid email${NC}\n"
fi

# Test 9: Validation - Short name
echo -e "${BLUE}Test 9: Validation test - Short name${NC}"
RESPONSE=$(curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"A","email":"test@example.com"}')
echo "Response: $RESPONSE"
if echo $RESPONSE | grep -q "error"; then
  echo -e "${GREEN}✓ Correctly validates name length${NC}\n"
else
  echo -e "${RED}✗ Should reject short name${NC}\n"
fi

# Test 10: Invalid endpoint
echo -e "${BLUE}Test 10: Invalid endpoint test${NC}"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X GET $BASE_URL/invalid)
echo "HTTP Status Code: $HTTP_CODE"
if [ "$HTTP_CODE" = "404" ]; then
  echo -e "${GREEN}✓ Correctly returns 404 for invalid endpoint${NC}\n"
else
  echo -e "${RED}✗ Should return 404 for invalid endpoint${NC}\n"
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}All tests completed!${NC}"
echo -e "${BLUE}========================================${NC}"
