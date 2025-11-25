# Enrollment System Architecture

## System Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    ENROLLMENT SYSTEM                             │
└─────────────────────────────────────────────────────────────────┘

┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Student │     │  Teacher │     │  Course  │     │ Database │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │                │                │                │
     
STUDENT FLOW:
     │
     │ 1. Browse Courses
     ├──────────────────────────────────►│
     │                                    │
     │ 2. GET /courses                    │
     │◄───────────────────────────────────┤
     │                                    │
     │ 3. Check Enrollment Status         │
     ├──────────────────────────────────►│
     │                                    │
     │ 4. GET /courses/:id/enrollment-status
     │◄───────────────────────────────────┤
     │                                    │
     │ 5. Enroll in Course                │
     ├──────────────────────────────────►│
     │                                    │
     │    Validate: Published? Own Course?│
     │                                    ├──────────►│
     │                                    │           │
     │                                    │  Create   │
     │                                    │  Enrollment
     │                                    │◄──────────┤
     │                                    │           │
     │ 6. Success Response                │           │
     │◄───────────────────────────────────┤           │
     │                                    │           │
     │ 7. View My Courses                 │           │
     ├──────────────────────────────────►│           │
     │                                    │           │
     │ 8. GET /enrollments/my-courses     │           │
     │                                    ├──────────►│
     │                                    │           │
     │                                    │  SELECT   │
     │                                    │  WHERE    │
     │                                    │  user_id  │
     │                                    │◄──────────┤
     │                                    │           │
     │ 9. List of Enrolled Courses        │           │
     │◄───────────────────────────────────┤           │
     │                                    │           │

TEACHER FLOW:
     │                │                   │           │
     │                │ 1. View Students  │           │
     │                ├──────────────────►│           │
     │                │                   │           │
     │                │ GET /courses/:id/enrollments  │
     │                │                   │           │
     │                │  Validate: Own Course?        │
     │                │                   ├──────────►│
     │                │                   │           │
     │                │                   │  SELECT   │
     │                │                   │  WHERE    │
     │                │                   │  course_id│
     │                │                   │◄──────────┤
     │                │                   │           │
     │                │ 2. Student List   │           │
     │                │◄──────────────────┤           │
     │                │                   │           │
```

## Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         HANDLER LAYER                            │
│  - EnrollmentHandler                                             │
│  - Enroll(), Unenroll(), GetMyEnrollments()                     │
│  - GetCourseEnrollments(), GetEnrollmentStatus()                │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                         SERVICE LAYER                            │
│  - EnrollmentService                                             │
│  - Business Logic                                                │
│  - Validation Rules                                              │
│  - Authorization Checks                                          │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                       REPOSITORY LAYER                           │
│  - EnrollmentRepository                                          │
│  - CRUD Operations                                               │
│  - Database Queries                                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│                         DATABASE LAYER                           │
│  - PostgreSQL                                                    │
│  - enrollments table                                             │
│  - Relationships: users, courses                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow

```
REQUEST → Middleware (Auth) → Handler → Service → Repository → Database
                    ↓
              Validate Token
              Check Role
                    ↓
              Extract User ID
                    ↓
                 Handler
              Parse Request
              Call Service
                    ↓
                 Service
           Validate Business Rules
           - Is course published?
           - Already enrolled?
           - Owns course?
                    ↓
               Repository
            Execute SQL Query
            Return Result
                    ↓
DATABASE ← Repository ← Service ← Handler ← Response
```

## Enrollment State Machine

```
┌─────────┐
│  START  │
└────┬────┘
     │
     │ Student enrolls
     ▼
┌─────────────┐
│   ACTIVE    │◄─────────┐
└─────┬───┬───┘          │
      │   │              │
      │   │ Unenroll     │ Re-enroll
      │   ▼              │
      │ ┌─────────────┐  │
      │ │   DROPPED   │──┘
      │ └─────────────┘
      │
      │ Progress = 100%
      ▼
┌─────────────┐
│  COMPLETED  │
└─────────────┘
```

## API Route Structure

```
/api/v1
├── /enrollments
│   ├── GET    /my-courses              (Get my enrolled courses)
│   └── PUT    /:enrollment_id/progress (Update progress)
│
└── /courses/:course_id
    ├── POST   /enroll                  (Enroll in course)
    ├── POST   /unenroll                (Unenroll from course)
    ├── GET    /enrollment-status       (Check enrollment status)
    └── GET    /enrollments             (View enrolled students - Teacher)
```

## Database Schema

```
┌─────────────────────────────────────────┐
│            enrollments                   │
├─────────────────────────────────────────┤
│ id (PK)                  SERIAL          │
│ user_id (FK)             INTEGER         │
│ course_id (FK)           INTEGER         │
│ status                   VARCHAR(20)     │
│ enrolled_at              TIMESTAMP       │
│ completed_at             TIMESTAMP       │
│ last_accessed_at         TIMESTAMP       │
│ progress_percent         FLOAT           │
│ created_at               TIMESTAMP       │
│ updated_at               TIMESTAMP       │
│ deleted_at               TIMESTAMP       │
└─────────────────────────────────────────┘
        │                    │
        │                    │
        ▼                    ▼
┌──────────────┐    ┌──────────────┐
│    users     │    │   courses    │
├──────────────┤    ├──────────────┤
│ id (PK)      │    │ id (PK)      │
│ name         │    │ title        │
│ email        │    │ description  │
│ role         │    │ teacher_id   │
│ ...          │    │ is_published │
└──────────────┘    └──────────────┘
```

## Authorization Matrix

```
┌──────────────┬─────────┬─────────┬───────┐
│   Action     │ Student │ Teacher │ Admin │
├──────────────┼─────────┼─────────┼───────┤
│ Enroll       │    ✓    │    ✓*   │   ✓   │
│ Unenroll     │    ✓    │    ✓    │   ✓   │
│ My Courses   │    ✓    │    ✓    │   ✓   │
│ View Students│    ✗    │   ✓**   │   ✓   │
│ Update Prog. │    ✓    │    ✓    │   ✓   │
└──────────────┴─────────┴─────────┴───────┘

 * Teachers can enroll in OTHER teacher's courses
** Teachers can only view THEIR OWN course enrollments
```

## Error Handling Flow

```
Request
   │
   ▼
Authentication Check
   │
   ├─ No Token ──────► 401 Unauthorized
   ├─ Invalid Token ─► 401 Unauthorized
   ├─ Blacklisted ───► 401 Token Revoked
   │
   ▼
Authorization Check
   │
   ├─ Wrong Role ────► 403 Forbidden
   │
   ▼
Validation
   │
   ├─ Invalid Course ID ──► 400 Bad Request
   ├─ Course Not Found ───► 404 Not Found
   ├─ Not Published ──────► 400 Not Published
   ├─ Already Enrolled ───► 409 Conflict
   ├─ Own Course ─────────► 400 Cannot Enroll
   │
   ▼
Database Operation
   │
   ├─ DB Error ───────────► 500 Internal Error
   │
   ▼
Success ──────────────────► 200/201 Success
```

## Summary

This enrollment system provides:
- ✅ Complete student enrollment workflow
- ✅ Teacher management interface
- ✅ Progress tracking
- ✅ Role-based access control
- ✅ Comprehensive error handling
- ✅ Clean architecture with separation of concerns

