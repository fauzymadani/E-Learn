# API Examples

## Authentication Endpoints

### Register
```
POST /api/v1/auth/register
```

**Request Body (JSON):**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "role": "student"
}
```

**Roles:** `student`, `teacher`, `admin`

**Success Response (201):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "student"
  }
}
```

### Login
```
POST /api/v1/auth/login
```

**Request Body (JSON):**
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Success Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "student"
  }
}
```

### Get Profile
```
GET /api/v1/auth/me
```

**Headers:**
- `Authorization: Bearer YOUR_TOKEN_HERE`

**Success Response (200):**
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "role": "student"
}
```

### Logout
```
POST /api/v1/auth/logout
```

**Headers:**
- `Authorization: Bearer YOUR_TOKEN_HERE`

**Success Response (200):**
```json
{
  "message": "successfully logged out"
}
```

**Example with cURL:**
```bash
curl -X POST "http://localhost:8080/api/v1/auth/logout" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Note:** For JWT-based authentication, the actual token invalidation happens on the client side by deleting/clearing the stored token. The logout endpoint logs the logout event on the server and can be extended to implement token blacklisting if needed.

---

## Create Lesson

### Endpoint
```
POST /api/v1/courses/:course_id/lessons
```

### Important Notes
- **Do NOT include trailing slash** - Use `/api/v1/courses/2/lessons` NOT `/api/v1/courses/2/lessons/`
- Content-Type: `multipart/form-data`
- Authentication required (Teacher/Admin role)

### Request Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| title | text | Yes | Lesson title |
| content | text | No | Lesson content/description |
| video | file | No | Video file (.mp4 or .mov) |
| file | file | No | PDF file (.pdf) |

### Example using cURL

```bash
# Basic lesson with just title and content
curl -X POST "http://localhost:8080/api/v1/courses/2/lessons" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -F "title=Introduction to Go" \
  -F "content=This lesson covers the basics of Go programming"

# Lesson with video
curl -X POST "http://localhost:8080/api/v1/courses/2/lessons" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -F "title=Introduction to Go" \
  -F "content=This lesson covers the basics of Go programming" \
  -F "video=@/path/to/video.mp4"

# Lesson with PDF file
curl -X POST "http://localhost:8080/api/v1/courses/2/lessons" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -F "title=Introduction to Go" \
  -F "content=This lesson covers the basics of Go programming" \
  -F "file=@/path/to/document.pdf"

# Lesson with both video and PDF
curl -X POST "http://localhost:8080/api/v1/courses/2/lessons" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -F "title=Introduction to Go" \
  -F "content=This lesson covers the basics of Go programming" \
  -F "video=@/path/to/video.mp4" \
  -F "file=@/path/to/document.pdf"
```

### Example using Postman/Thunder Client/REST Client

1. Set request method to `POST`
2. Set URL to `http://localhost:8080/api/v1/courses/2/lessons` (NO trailing slash)
3. In Headers, add:
   - `Authorization: Bearer YOUR_TOKEN_HERE`
4. In Body:
   - Select `form-data` (NOT `x-www-form-urlencoded`)
   - Add fields:
     - `title` (Text): "Introduction to Go"
     - `content` (Text): "This lesson covers the basics"
     - `video` (File): Select your .mp4 or .mov file (optional)
     - `file` (File): Select your .pdf file (optional)

### Common Errors

#### Error: "multipart: NextPart: EOF"
**Cause**: Request has multipart/form-data content-type but empty or malformed body
**Solution**: 
- Make sure you're sending form-data, not raw JSON
- Ensure at least the `title` field is included
- Remove trailing slash from URL

#### Error: "redirecting request 307"
**Cause**: Using trailing slash in URL which causes redirect and corrupts multipart data
**Solution**: Remove trailing slash - use `/lessons` not `/lessons/`

#### Error: "invalid course id"
**Cause**: Course ID in URL is not a valid number
**Solution**: Use numeric course ID like `/courses/2/lessons`

#### Error: "title is required"
**Cause**: Missing or empty title field
**Solution**: Include `title` field in form-data

#### PDF/Video not uploading but no error shown
**Possible causes**:
1. **Wrong field name**: Use `file` for PDF, `video` for video (not `pdf` or other names)
2. **File extension case**: Use lowercase `.pdf` not `.PDF` (now fixed - case-insensitive)
3. **Content-Type manually set**: Don't set Content-Type header manually, let the client set it
4. **Permissions**: Check if uploads/files and uploads/videos directories are writable
5. **Disk space**: Ensure there's enough disk space

**How to debug**:
- Check server logs for "Received file field" and "PDF/Video uploaded successfully" messages
- Run the test script: `./test_pdf_upload.sh`
- Check uploads directory: `ls -la uploads/files/` and `ls -la uploads/videos/`
- Verify file permissions: `chmod -R 755 uploads/`

### Success Response
```json
{
  "id": 1,
  "course_id": 2,
  "title": "Introduction to Go",
  "content": "This lesson covers the basics of Go programming",
  "video_url": "/uploads/videos/2_1700912345.mp4",
  "file_url": "/uploads/files/2_1700912345.pdf",
  "order": 1,
  "created_at": "2025-11-25T11:39:10Z",
  "updated_at": "2025-11-25T11:39:10Z"
}
```

## Other Lesson Endpoints

### Get All Lessons for a Course
```
GET /api/v1/courses/:course_id/lessons
```
No authentication required. Returns all lessons for the specified course.

### Get Single Lesson
```
GET /api/v1/courses/:course_id/lessons/:lesson_id
```
No authentication required.

### Update Lesson
```
PUT /api/v1/courses/:course_id/lessons/:lesson_id
```
Authentication required (Teacher/Admin). Uses same multipart/form-data format as Create.

### Delete Lesson
```
DELETE /api/v1/courses/:course_id/lessons/:lesson_id
```
Authentication required (Teacher/Admin).

### Reorder Lessons
```
PUT /api/v1/courses/:course_id/lessons/reorder
```
Authentication required (Teacher/Admin).

Body (JSON):
```json
{
  "lesson_orders": [
    {"lesson_id": 1, "order": 2},
    {"lesson_id": 2, "order": 1},
    {"lesson_id": 3, "order": 3}
  ]
}
```

---

## Notifications

### Get All Notifications
```
GET /api/v1/notifications
```
Authentication required.

Query Parameters:
- `page` (optional, default: 1) - Page number
- `limit` (optional, default: 20) - Items per page
- `unread_only` (optional, default: false) - Set to "true" to get only unread notifications

Example:
```bash
curl -X GET "{{BASE_URL}}/v1/notifications?page=1&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:
```json
{
  "notifications": [
    {
      "id": 1,
      "user_id": 7,
      "type": "enrollment",
      "title": "New Student Enrolled",
      "message": "A student has enrolled in your course: Introduction to Programming",
      "is_read": false,
      "created_at": "2025-11-25T12:00:00Z"
    },
    {
      "id": 2,
      "user_id": 7,
      "type": "completed",
      "title": "Course Completed",
      "message": "Congratulations! You have completed the course: Web Development",
      "is_read": false,
      "created_at": "2025-11-25T11:00:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "limit": 20
}
```

### Get Unread Count
```
GET /api/v1/notifications/unread-count
```
Authentication required.

Example:
```bash
curl -X GET "{{BASE_URL}}/v1/notifications/unread-count" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:
```json
{
  "unread_count": 5
}
```

### Mark Notification as Read
```
PUT /api/v1/notifications/:id/read
```
Authentication required.

Example:
```bash
curl -X PUT "{{BASE_URL}}/v1/notifications/1/read" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:
```json
{
  "message": "notification marked as read"
}
```

### Mark All Notifications as Read
```
PUT /api/v1/notifications/read-all
```
Authentication required.

Example:
```bash
curl -X PUT "{{BASE_URL}}/v1/notifications/read-all" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:
```json
{
  "message": "all notifications marked as read",
  "count": 5
}
```
