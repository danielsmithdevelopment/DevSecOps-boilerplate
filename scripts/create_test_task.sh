#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCIsImV4cCI6MTc0NDUyNTEwNCwiaWF0IjoxNzQ0NDM4NzA0fQ.8Q9Pq7-ljdkvbjfGH5ZFhH72u_M4s8gBPqGxxUYfiq8"

# Task configuration
TASK_NAME="test-http-task"
TASK_TYPE="http"
TASK_SCHEDULE="*/5 * * * *"  # Run every 5 minutes
TASK_CONFIG='{
  "http": {
    "url": "https://httpbin.org/get",
    "method": "GET",
    "headers": {
      "User-Agent": "TaskRunner/1.0"
    }
  }
}'

# Create task
echo "Creating test task..."
curl -X POST "${API_URL}/task/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -d "{
    \"name\": \"${TASK_NAME}\",
    \"type\": \"${TASK_TYPE}\",
    \"schedule\": \"${TASK_SCHEDULE}\",
    \"config\": ${TASK_CONFIG}
  }"

echo -e "\n\nTask created. You can check its status using:"
echo "curl -X GET \"${API_URL}/task/status\" -H \"Authorization: Bearer ${JWT_TOKEN}\"" 