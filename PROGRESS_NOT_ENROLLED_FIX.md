# Progress Tracking - "User Not Enrolled" Error Fix

## Error

```json
{
  "error": "user not enrolled in this course"
}
```

When calling: `POST {{BASE_URL}}/v1/progress/lessons/:lesson_id/complete`

## Root Cause

The progress tracking system requires you to be **enrolled in the course** before you can mark lessons as complete. This is a security feature to ensure only enrolled students can track progress.

## Solution: Enroll First!

### Step 1: Find the Course ID

Get the lesson details to find which course it belongs to:

```bash
curl -X GET "http://localhost:8080/api/v1/courses/:course_id/lessons/:lesson_id" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Step 2: Enroll in the Course

```bash
curl -X POST "http://localhost:8080/api/v1/courses/:course_id/enroll" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Success Response:**
```json
{
  "id": 1,
  "user_id": 7,
  "course_id": 2,
  "status": "active",
  "enrolled_at": "2025-11-25T13:00:00Z",
  "progress_percent": 0
}
```

### Step 3: Now Mark Lesson as Complete

```bash
curl -X POST "http://localhost:8080/api/v1/progress/lessons/:lesson_id/complete" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Success Response:**
```json
{
  "message": "lesson marked as completed",
  "lesson_id": 1
}
```

## Complete Workflow Example

```bash
# 1. Login
TOKEN=$(curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"student@example.com","password":"password123"}' \
  | jq -r '.access_token')

# 2. Get lesson details (to find course_id)
LESSON_DETAILS=$(curl -X GET "http://localhost:8080/api/v1/courses/2/lessons/1" \
  -H "Authorization: Bearer $TOKEN")

echo "Lesson details:"
echo $LESSON_DETAILS | jq '.'

# Extract course_id
COURSE_ID=$(echo $LESSON_DETAILS | jq -r '.course_id')
echo "Course ID: $COURSE_ID"

# 3. Check if already enrolled
ENROLLMENT_STATUS=$(curl -s -X GET "http://localhost:8080/api/v1/courses/${COURSE_ID}/enrollment-status" \
  -H "Authorization: Bearer $TOKEN")

IS_ENROLLED=$(echo $ENROLLMENT_STATUS | jq -r '.enrolled')

if [ "$IS_ENROLLED" != "true" ]; then
    echo "Not enrolled, enrolling now..."
    
    # 4. Enroll in course
    curl -X POST "http://localhost:8080/api/v1/courses/${COURSE_ID}/enroll" \
      -H "Authorization: Bearer $TOKEN"
    
    echo "✓ Enrolled successfully"
else
    echo "✓ Already enrolled"
fi

# 5. Now mark lesson as complete
curl -X POST "http://localhost:8080/api/v1/progress/lessons/1/complete" \
  -H "Authorization: Bearer $TOKEN"

echo "✓ Lesson marked as complete"

# 6. Check course progress
curl -X GET "http://localhost:8080/api/v1/progress/courses/${COURSE_ID}" \
  -H "Authorization: Bearer $TOKEN"
```

## Why This Requirement?

The enrollment check prevents:
- ❌ Random users from marking lessons complete in courses they don't own
- ❌ Progress tracking for courses you haven't joined
- ❌ Unauthorized access to course content
- ✅ Ensures only enrolled students can track progress
- ✅ Maintains data integrity

## Check Your Enrollment Status

```bash
# Check if enrolled in a specific course
curl -X GET "http://localhost:8080/api/v1/courses/:course_id/enrollment-status" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response if enrolled:**
```json
{
  "enrolled": true,
  "enrollment": {
    "id": 1,
    "status": "active",
    "progress_percent": 0
  }
}
```

**Response if NOT enrolled:**
```json
{
  "enrolled": false,
  "message": "not enrolled in this course"
}
```

## Quick Fix

If you just want to test the progress feature, enroll in the course first:

```bash
# Replace :course_id with actual course ID (e.g., 2)
curl -X POST "http://localhost:8080/api/v1/courses/2/enroll" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Then try marking the lesson as complete again!

## Progress Tracking Endpoints

Once enrolled, you can use:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/progress/lessons/:lesson_id/complete` | Mark lesson complete |
| DELETE | `/api/v1/progress/lessons/:lesson_id/complete` | Unmark lesson |
| GET | `/api/v1/progress/courses/:course_id` | Get course progress |
| GET | `/api/v1/progress/lessons/:lesson_id` | Get lesson progress |

## Summary

✅ **Fixed:** Removed CompletedAt field reference (doesn't exist in DB)
⚠️ **Requirement:** Must enroll in course before marking lessons complete
📝 **Solution:** Enroll using `POST /api/v1/courses/:course_id/enroll`

## Test Script

Save this as `test_progress.sh`:

```bash
#!/bin/bash
BASE_URL="http://localhost:8080"

# Login
echo "Logging in..."
TOKEN=$(curl -s -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"student@test.com","password":"password123"}' \
  | jq -r '.access_token')

COURSE_ID=2
LESSON_ID=1

# Check enrollment
echo "Checking enrollment..."
STATUS=$(curl -s -X GET "${BASE_URL}/api/v1/courses/${COURSE_ID}/enrollment-status" \
  -H "Authorization: Bearer $TOKEN")

IS_ENROLLED=$(echo $STATUS | jq -r '.enrolled')

if [ "$IS_ENROLLED" != "true" ]; then
    echo "Enrolling in course..."
    curl -s -X POST "${BASE_URL}/api/v1/courses/${COURSE_ID}/enroll" \
      -H "Authorization: Bearer $TOKEN" | jq '.'
fi

# Mark lesson complete
echo "Marking lesson as complete..."
curl -s -X POST "${BASE_URL}/api/v1/progress/lessons/${LESSON_ID}/complete" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# Check progress
echo "Getting course progress..."
curl -s -X GET "${BASE_URL}/api/v1/progress/courses/${COURSE_ID}" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

Run with: `chmod +x test_progress.sh && ./test_progress.sh`

