# API Quick Reference Guide

คู่มืออ้างอิงด่วนสำหรับ API endpoints ที่ใช้บ่อย

## Authentication

```bash
# Login
GET /api/auth/google

# Get Profile
GET /api/profile
Headers: Authorization: Bearer <token>, X-API-Key: <key>

# Refresh Token
POST /api/auth/refresh
Body: { "refresh_token": "..." }
```

## Courses

```bash
# List Courses
GET /api/courses?status=active&page=1&limit=10

# Create Course
POST /api/courses
Body: { "name": "...", "code": "...", "description": "..." }

# Get Course
GET /api/courses/:id

# Update Course
PUT /api/courses/:id
Body: { "name": "...", ... }

# Delete Course
DELETE /api/courses/:id
```

## Enrollment

```bash
# Enroll
POST /api/courses/:id/enroll
Body: { "role": "student" }

# Get Enrollments
GET /api/courses/:id/enrollments

# Unenroll
DELETE /api/courses/:id/enroll
```

## Course Materials

```bash
# List Materials
GET /api/course-materials?course_id=1&type=code_exercise

# Create Material
POST /api/course-materials
Body: {
  "course_id": 1,
  "type": "code_exercise",
  "title": "...",
  "description": "...",
  "deadline": "2024-12-31T23:59:59Z"
}

# Get Material
GET /api/course-materials/:id

# Update Material
PUT /api/course-materials/:id

# Delete Material
DELETE /api/course-materials/:id

# Upload File
POST /api/course-materials/upload
Form Data: file, course_id, type
```

## Test Cases

```bash
# Get Test Cases
GET /api/course-materials/:id/test-cases

# Add Test Case
POST /api/course-materials/:id/test-cases
Body: {
  "input": "...",
  "expected_output": "...",
  "is_hidden": false
}

# Update Test Case
PUT /api/course-materials/test-cases/:test_case_id

# Delete Test Case
DELETE /api/course-materials/test-cases/:test_case_id
```

## Submissions

```bash
# Submit Exercise
POST /api/course-materials/:id/submit
Body: {
  "code": "def solution():\n    return 42",
  "language": "python"
}

# Get My Submission
GET /api/course-materials/:id/submissions/me

# Get Submission
GET /api/submissions/:id

# Submit PDF
POST /api/course-materials/:id/submit-pdf
Form Data: file
```

## Progress

```bash
# Get Self Progress
GET /api/students/progress?course_id=1

# Get Course Progress
GET /api/courses/:id/progress

# Verify Progress
POST /api/progress/:id/verify
Body: {
  "status": "approved",
  "feedback": "..."
}
```

## Announcements

```bash
# List Announcements
GET /api/announcements?course_id=1

# Create Announcement
POST /api/announcements
Body: {
  "course_id": 1,
  "title": "...",
  "content": "..."
}

# Get Announcement
GET /api/announcements/:id

# Update Announcement
PUT /api/announcements/:id

# Delete Announcement
DELETE /api/announcements/:id
```

## Queue

```bash
# Get Jobs
GET /api/queue/jobs?status=pending

# Get Job
GET /api/queue/jobs/:id

# Cancel Job
POST /api/queue/jobs/:id/cancel

# Get Stats
GET /api/queue/stats
```

## Drafts

```bash
# Save Draft
POST /api/drafts/exercises/:exercise_id
Body: { "code": "..." }

# Get Draft
GET /api/drafts/exercises/:exercise_id

# Delete Draft
DELETE /api/drafts/exercises/:exercise_id
```

## Deadline Checker

```bash
# Check Deadline
GET /api/materials/check-deadline?material_id=1

# Get Available Materials
GET /api/materials/available?course_id=1

# Get Expired Materials
GET /api/materials/expired?course_id=1

# Can Submit
GET /api/materials/can-submit?exercise_id=1

# Upcoming Deadlines
GET /api/materials/upcoming-deadlines?course_id=1&hours=24
```

## Playground (FastAPI)

```bash
# Run Code
POST /api/playground/run
Headers: X-API-Key: <key>
Body: {
  "code": "class Node: ...",
  "dataType": "singlylinkedlist"
}

# Health Check
GET /health
```

## PDF Exercises

```bash
# Get My PDF Submission
GET /api/materials/:material_id/submissions/me

# Get PDF Submissions
GET /api/materials/:material_id/submissions

# Approve Submission
POST /api/submissions/:submission_id/approve
Body: { "feedback": "...", "score": 100 }

# Reject Submission
POST /api/submissions/:submission_id/reject
Body: { "feedback": "...", "reason": "..." }
```

## Common Headers

```bash
# Required for most endpoints
Authorization: Bearer <jwt-token>
X-API-Key: <api-key>
Content-Type: application/json
```

## Material Types

- `document` - เอกสาร (PDF, DOC, etc.)
- `video` - วิดีโอ (YouTube, Vimeo, etc.)
- `code_exercise` - แบบฝึกหัดโค้ด
- `pdf_exercise` - แบบฝึกหัด PDF

## Submission Status

- `pending` - กำลังรอ
- `processing` - กำลังประมวลผล
- `completed` - เสร็จสิ้น
- `failed` - ล้มเหลว
- `graded` - ตรวจแล้ว

## Progress Status

- `pending` - กำลังรอ
- `submitted` - ส่งแล้ว
- `approved` - อนุมัติแล้ว
- `rejected` - ปฏิเสธแล้ว

---

*Quick Reference - Last Updated: 2024*












