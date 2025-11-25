# 📚 Enrollment System - Quick Reference

## ✅ What Was Implemented

Complete enrollment system for students to enroll in courses and teachers to view enrolled students.

## 🎯 Key Features

### Student Features
- ✅ **Enroll in courses** - Join published courses
- ✅ **View My Courses** - See all enrolled courses
- ✅ **Track Progress** - Monitor completion (0-100%)
- ✅ **Check Status** - Verify enrollment in specific courses
- ✅ **Unenroll** - Drop courses if needed

### Teacher Features
- ✅ **View Enrolled Students** - See who's in their courses
- ✅ **Track Student Progress** - Monitor individual progress
- ✅ **Enrollment Details** - View dates and status

## 📍 API Endpoints Summary

| Method | Endpoint | Auth | Role | Description |
|--------|----------|------|------|-------------|
| POST | `/api/v1/courses/:course_id/enroll` | ✓ | Student/Admin | Enroll in course |
| POST | `/api/v1/courses/:course_id/unenroll` | ✓ | Any | Unenroll from course |
| GET | `/api/v1/enrollments/my-courses` | ✓ | Any | Get my enrolled courses |
| GET | `/api/v1/courses/:course_id/enrollments` | ✓ | Teacher/Admin | View enrolled students |
| GET | `/api/v1/courses/:course_id/enrollment-status` | ✓ | Any | Check enrollment status |
| PUT | `/api/v1/enrollments/:id/progress` | ✓ | Any | Update progress |

## 🚀 Quick Start

### Student: Enroll in a Course

```bash
# 1. Login
TOKEN=$(curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"student@example.com","password":"password"}' \
  | jq -r '.access_token')

# 2. Enroll
curl -X POST "http://localhost:8080/api/v1/courses/1/enroll" \
  -H "Authorization: Bearer $TOKEN"

# 3. View my courses
curl -X GET "http://localhost:8080/api/v1/enrollments/my-courses" \
  -H "Authorization: Bearer $TOKEN"
```

### Teacher: View Enrolled Students

```bash
# Login as teacher
TEACHER_TOKEN=$(curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"teacher@example.com","password":"password"}' \
  | jq -r '.access_token')

# View enrollments
curl -X GET "http://localhost:8080/api/v1/courses/1/enrollments" \
  -H "Authorization: Bearer $TEACHER_TOKEN"
```

## 📊 Enrollment Statuses

- **active** - Currently enrolled
- **completed** - Finished (100% progress)
- **dropped** - Unenrolled

## 🔒 Business Rules

1. ✅ Only students/admins can enroll
2. ✅ Only published courses can be enrolled in
3. ✅ Teachers cannot enroll in their own courses
4. ✅ No duplicate enrollments
5. ✅ Teachers can only view their own course enrollments

## 🧪 Testing

Run the test script:
```bash
chmod +x test_enrollment.sh
./test_enrollment.sh
```

## 📁 Files Created

1. `internal/domain/enrollment.go` - Data model
2. `internal/repository/enrollment.go` - Database layer
3. `internal/service/enrollment.go` - Business logic
4. `internal/handler/enrollment.go` - API handlers
5. `internal/router/router.go` - Routes (updated)
6. `cmd/api/main.go` - Initialization (updated)

## 📖 Full Documentation

See `ENROLLMENT_SYSTEM.md` for complete documentation including:
- Detailed API examples
- Error handling
- Integration guides
- Database schema
- Advanced usage

## ✨ Example Responses

### My Courses
```json
[
  {
    "id": 1,
    "user_id": 5,
    "course_id": 1,
    "status": "active",
    "enrolled_at": "2025-11-25T10:00:00Z",
    "progress_percent": 45.5,
    "course": {
      "id": 1,
      "title": "Introduction to Programming"
    }
  }
]
```

### Enrolled Students (Teacher View)
```json
[
  {
    "id": 1,
    "user_id": 5,
    "course_id": 1,
    "status": "active",
    "progress_percent": 45.5,
    "user": {
      "id": 5,
      "name": "John Doe",
      "email": "john@example.com"
    }
  }
]
```

## 🎉 Ready to Use!

The enrollment system is fully implemented and ready for use. Students can now enroll in courses and teachers can track their students!

