#!/bin/bash

# Colors for Terminal
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "=========================================="
echo "   HOTLINES3 API TESTING SCRIPT"
echo "=========================================="

# Variables
BASE_URL="http://localhost:8080/v1"
REGISTER_URL="$BASE_URL/auth/register"
LOGIN_URL="$BASE_URL/auth/login"

# 1. Test Register
echo -e "\n${GREEN}[TEST 1] Register User${NC}"
echo "Username: 123456"
echo "Password: password123"

curl -X POST $REGISTER_URL \
  -H "Content-Type: application/json" \
  -d '{
    "username": "123456",
    "password": "password123",
    "role": "admin"
  }'

echo -e "\n\n----------------------------------------"

# 2. Test Register Duplicate (Should fail)
echo -e "\n${GREEN}[TEST 2] Register Duplicate User (Should Fail)${NC}"
curl -X POST $REGISTER_URL \
  -H "Content-Type: application/json" \
  -d '{
    "username": "123456",
    "password": "password123",
    "role": "user"
  }'

echo -e "\n\n----------------------------------------"

# 3. Test Login (Wrong Password)
echo -e "\n${GREEN}[TEST 3] Login Wrong Password (Should Fail)${NC}"
curl -X POST $LOGIN_URL \
  -H "Content-Type: application/json" \
  -d '{
    "username": "123456",
    "password": "wrongpass"
  }'

echo -e "\n\n----------------------------------------"

# 4. Test Login (Success)
echo -e "\n${GREEN}[TEST 4] Login Success${NC}"
echo "Expecting Token in response..."

RESPONSE=$(curl -s -X POST $LOGIN_URL \
  -H "Content-Type: application/json" \
  -d '{
    "username": "123456",
    "password": "password123"
  }')

echo "$RESPONSE" | jq '.'

# Extract Token from response
TOKEN=$(echo $RESPONSE | jq -r '.data.accessToken')

echo -e "\n\n----------------------------------------"

# 5. Verify Token
if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
    echo -e "\n${GREEN}[TEST 5] Token Generated Successfully${NC}"
    echo "Token (truncated): ${TOKEN:0:20}..."
else
    echo -e "\n${GREEN}[TEST 5] Failed to generate Token${NC}"
fi

echo -e "\n=========================================="
echo "   TEST COMPLETED"
echo "=========================================="
