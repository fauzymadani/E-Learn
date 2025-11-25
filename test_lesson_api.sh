#!/bin/bash

# Test script for creating a lesson
# Make sure the server is running first: go run cmd/api/main.go

# Configuration
BASE_URL="http://localhost:8080"
COURSE_ID=2

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Testing Lesson Creation API${NC}"
echo "================================"
echo ""

# First, you need to login and get a token
echo -e "${YELLOW}Step 1: Login to get authentication token${NC}"
echo "If you don't have a teacher/admin account, register one first"
echo ""
echo "Example:"
echo "curl -X POST \"${BASE_URL}/api/v1/auth/login\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"email\":\"teacher@example.com\",\"password\":\"password123\"}'"
echo ""
read -p "Enter your auth token (from login response): " TOKEN
echo ""

# Test 1: Create lesson with just title and content
echo -e "${YELLOW}Test 1: Creating lesson with title and content${NC}"
RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/courses/${COURSE_ID}/lessons" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "title=Introduction to API Testing" \
  -F "content=This lesson covers how to test REST APIs")

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" -eq 201 ]; then
    echo -e "${GREEN}✓ Success! Lesson created${NC}"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ Failed with status: $HTTP_STATUS${NC}"
    echo "$BODY"
fi
echo ""

# Test 2: Create lesson without title (should fail)
echo -e "${YELLOW}Test 2: Creating lesson without title (should fail)${NC}"
RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/courses/${COURSE_ID}/lessons" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "content=This should fail")

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$RESPONSE" | sed '/HTTP_STATUS:/d')

if [ "$HTTP_STATUS" -eq 400 ]; then
    echo -e "${GREEN}✓ Correctly rejected (400 Bad Request)${NC}"
    echo "$BODY"
else
    echo -e "${RED}✗ Unexpected status: $HTTP_STATUS${NC}"
    echo "$BODY"
fi
echo ""

# Test 3: Using wrong URL (with trailing slash)
echo -e "${YELLOW}Test 3: Using URL with trailing slash (OLD way - should cause issues)${NC}"
RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/api/v1/courses/${COURSE_ID}/lessons/" \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "title=Test with trailing slash" \
  -F "content=This might cause redirect")

HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d: -f2)

if [ "$HTTP_STATUS" -eq 307 ] || [ "$HTTP_STATUS" -eq 301 ]; then
    echo -e "${RED}✗ Redirected (${HTTP_STATUS}) - This is the problem!${NC}"
    echo "Always use URL WITHOUT trailing slash"
elif [ "$HTTP_STATUS" -eq 201 ]; then
    echo -e "${YELLOW}Note: It worked, but prefer URL without trailing slash${NC}"
else
    echo -e "${RED}Status: $HTTP_STATUS${NC}"
fi
echo ""

echo -e "${GREEN}Testing complete!${NC}"
echo ""
echo "Summary:"
echo "- Always use: /api/v1/courses/:course_id/lessons (NO trailing slash)"
echo "- Use multipart/form-data format with -F flag in curl"
echo "- Include Authorization header with Bearer token"
echo "- Title field is required"

