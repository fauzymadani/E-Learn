# Notification API Quick Reference

## Overview
The notification system allows users to receive and manage notifications about important events in the platform.

## Available Endpoints

### 1. Get All Notifications
**GET** `/api/v1/notifications`

Retrieve all notifications for the authenticated user.

**Query Parameters:**
- `page` (optional, default: 1) - Page number
- `limit` (optional, default: 20) - Number of items per page
- `unread_only` (optional) - Set to "true" to get only unread notifications

**Example:**
```bash
# Get all notifications (page 1)
curl -X GET "http://localhost:8080/api/v1/notifications" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get only unread notifications
curl -X GET "http://localhost:8080/api/v1/notifications?unread_only=true" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get page 2 with 10 items per page
curl -X GET "http://localhost:8080/api/v1/notifications?page=2&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### 2. Get Unread Count
**GET** `/api/v1/notifications/unread-count`

Get the count of unread notifications for the authenticated user.

**Example:**
```bash
curl -X GET "http://localhost:8080/api/v1/notifications/unread-count" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "unread_count": 5
}
```

---

### 3. Mark Notification as Read
**PUT** `/api/v1/notifications/:id/read`

Mark a specific notification as read.

**Example:**
```bash
curl -X PUT "http://localhost:8080/api/v1/notifications/1/read" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "message": "notification marked as read"
}
```

---

### 4. Mark All Notifications as Read
**PUT** `/api/v1/notifications/read-all`

Mark all notifications for the authenticated user as read.

**Example:**
```bash
curl -X PUT "http://localhost:8080/api/v1/notifications/read-all" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**
```json
{
  "message": "all notifications marked as read",
  "count": 5
}
```

---

## Notification Types

### 1. **enrollment**
Sent to teachers when a student enrolls in their course.

**Example:**
```json
{
  "id": 1,
  "user_id": 3,
  "type": "enrollment",
  "title": "New Student Enrolled",
  "message": "A student has enrolled in your course: Introduction to Programming",
  "is_read": false,
  "created_at": "2025-11-25T12:00:00Z"
}
```

### 2. **completed**
Sent to students when they complete all lessons in a course.

**Example:**
```json
{
  "id": 2,
  "user_id": 7,
  "type": "completed",
  "title": "Course Completed",
  "message": "Congratulations! You have completed the course: Web Development",
  "is_read": false,
  "created_at": "2025-11-25T12:30:00Z"
}
```

---

## Testing Workflow

### Test as a Teacher (Receiving Enrollment Notifications)

1. Login as a teacher
2. Create a course
3. Login as a student (different account)
4. Enroll in the teacher's course
5. Login back as the teacher
6. Check notifications:
```bash
curl -X GET "http://localhost:8080/api/v1/notifications" \
  -H "Authorization: Bearer TEACHER_TOKEN"
```

### Test as a Student (Receiving Completion Notifications)

1. Login as a student
2. Enroll in a course
3. Complete all lessons in the course
4. Check notifications:
```bash
curl -X GET "http://localhost:8080/api/v1/notifications" \
  -H "Authorization: Bearer STUDENT_TOKEN"
```

---

## Integration Requirements

**Important:** The notification gRPC service must be running on the configured address (default: `localhost:50051`).

If the notification service is not available:
- The API will still work normally
- Notifications won't be sent
- A warning will be logged: "Warning: notification service not available"

---

## Common Use Cases

### Dashboard Badge - Show Unread Count
```bash
GET /api/v1/notifications/unread-count
```

### Notification Center - List All Notifications
```bash
GET /api/v1/notifications?limit=50
```

### Mark Notification When User Clicks It
```bash
PUT /api/v1/notifications/:id/read
```

### Clear All Notifications Button
```bash
PUT /api/v1/notifications/read-all
```

---

## Error Responses

**500 Internal Server Error:**
```json
{
  "error": "failed to get notifications"
}
```

**401 Unauthorized:**
```json
{
  "error": "unauthorized"
}
```

**400 Bad Request:**
```json
{
  "error": "invalid notification id"
}
```

