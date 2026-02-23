# DSView Backend API Documentation

## สารบัญ

1. [ภาพรวม](#ภาพรวม)
2. [การ Authentication](#การ-authentication)
3. [Go Backend API](#go-backend-api)
4. [FastAPI Backend API](#fastapi-backend-api)
5. [Error Codes](#error-codes)

---

## ภาพรวม

ระบบ DSView Backend ประกอบด้วย 2 ส่วนหลัก:

1. **Go Backend** - API หลักสำหรับจัดการ Authentication, Courses, Materials, Submissions, และอื่นๆ
2. **FastAPI Backend** - API สำหรับ Code Execution และ Playground

### Base URLs

- **Go Backend**: `http://localhost:8080` (หรือตามที่กำหนดใน config)
- **FastAPI Backend**: `http://localhost:8000` (หรือตามที่กำหนดใน config)

### API Version

- Go Backend: `1.0.0`
- FastAPI Backend: `0.0.5-alpha`

---

## การ Authentication

### 1. API Key Authentication

ใช้สำหรับ endpoints ที่ต้องการความปลอดภัยสูง

**Header:**
```
X-API-Key: your-api-key
```

### 2. JWT Authentication

ใช้สำหรับ endpoints ที่ต้องการ authentication ของ user

**Header:**
```
Authorization: Bearer <jwt-token>
```

### 3. Combined Authentication

บาง endpoints ต้องการทั้ง API Key และ JWT Token

**Headers:**
```
X-API-Key: your-api-key
Authorization: Bearer <jwt-token>
```

### 4. OAuth 2.0 (Google)

สำหรับการ login ผ่าน Google

---

## Go Backend API

### Authentication & User Management

#### 1. Google OAuth Login
```
GET /api/auth/google
```
**Description:** เริ่มต้น OAuth flow สำหรับ Google login

**Authentication:** ไม่ต้องใช้

**Response:** Redirect ไปยัง Google OAuth page

---

#### 2. Google OAuth Callback
```
GET /api/auth/google/callback
```
**Description:** Callback endpoint สำหรับ Google OAuth

**Authentication:** ไม่ต้องใช้

**Query Parameters:**
- `code` (string, required) - Authorization code จาก Google
- `state` (string, optional) - State parameter

**Response:** Redirect ไปยัง frontend พร้อม JWT token

---

#### 3. Logout
```
POST /api/auth/logout
```
**Description:** Logout user และลบ session

**Authentication:** ไม่ต้องใช้ (แต่ควรส่ง JWT token ถ้ามี)

**Response:**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

---

#### 4. Refresh Token
```
POST /api/auth/refresh
```
**Description:** Refresh JWT token

**Authentication:** JWT (อนุญาตให้ใช้ expired token)

**Request Body:**
```json
{
  "refresh_token": "your-refresh-token"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "new-access-token",
    "refresh_token": "new-refresh-token",
    "expires_in": 3600
  }
}
```

---

#### 5. Get User Profile
```
GET /api/profile
GET /api/profile/me
```
**Description:** ดึงข้อมูล profile ของ user ที่ login

**Authentication:** JWT + API Key

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "picture": "https://...",
    "role": "student"
  }
}
```

---

### Courses

#### 1. List Courses
```
GET /api/courses
```
**Description:** ดึงรายการ courses ทั้งหมด

**Authentication:** API Key + JWT

**Query Parameters:**
- `status` (string, optional) - Filter by status: `active`, `archived`
- `page` (int, optional) - Page number
- `limit` (int, optional) - Items per page

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "Data Structures",
      "code": "CS101",
      "description": "...",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

#### 2. Create Course
```
POST /api/courses
```
**Description:** สร้าง course ใหม่

**Authentication:** API Key + JWT (Teacher/Admin only)

**Request Body:**
```json
{
  "name": "Data Structures",
  "code": "CS101",
  "description": "Course description",
  "status": "active"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Data Structures",
    "code": "CS101",
    ...
  }
}
```

---

#### 3. Get Course
```
GET /api/courses/:id
```
**Description:** ดึงข้อมูล course ตาม ID

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Course ID

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Data Structures",
    ...
  }
}
```

---

#### 4. Update Course
```
PUT /api/courses/:id
```
**Description:** อัปเดตข้อมูล course

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Course ID

**Request Body:**
```json
{
  "name": "Updated Name",
  "description": "Updated description"
}
```

---

#### 5. Delete Course
```
DELETE /api/courses/:id
```
**Description:** ลบ course

**Authentication:** API Key + JWT (Admin only)

**Path Parameters:**
- `id` (int, required) - Course ID

---

#### 6. Get Course Exercises
```
GET /api/courses/:id/exercises
```
**Description:** ดึงรายการ exercises ใน course

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Course ID

---

#### 7. Get Course Report (Teacher)
```
GET /api/courses/:id/report/teacher
```
**Description:** ดึงรายงาน course สำหรับ teacher

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Course ID

---

#### 8. Get Course Report (TA)
```
GET /api/courses/:id/report/ta
```
**Description:** ดึงรายงาน course สำหรับ TA

**Authentication:** API Key + JWT (TA/Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Course ID

---

#### 9. Delete Course Image
```
DELETE /api/courses/:id/image
```
**Description:** ลบรูปภาพของ course

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Course ID

---

### Course Enrollment

#### 1. Enroll in Course
```
POST /api/courses/:id/enroll
POST /api/enrollments/courses/:id
```
**Description:** ลงทะเบียนใน course

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Course ID

**Request Body:**
```json
{
  "role": "student"
}
```

---

#### 2. Get Course Enrollments
```
GET /api/courses/:id/enrollments
GET /api/enrollments/courses/:id
```
**Description:** ดึงรายการ enrollments ใน course

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `id` (int, required) - Course ID

---

#### 3. Get My Enrollment
```
GET /api/courses/:id/my-enrollment
```
**Description:** ดึงข้อมูล enrollment ของตัวเองใน course

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Course ID

---

#### 4. Unenroll from Course
```
DELETE /api/courses/:id/enroll
DELETE /api/enrollments/courses/:id
```
**Description:** ยกเลิกการลงทะเบียนใน course

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Course ID

---

### Course Materials

#### 1. List Course Materials
```
GET /api/course-materials
```
**Description:** ดึงรายการ materials ใน course

**Authentication:** API Key + JWT

**Query Parameters:**
- `course_id` (int, required) - Course ID
- `type` (string, optional) - Filter by type: `document`, `video`, `code_exercise`, `pdf_exercise`
- `week` (int, optional) - Filter by week number

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "course_id": 1,
      "type": "code_exercise",
      "title": "Exercise 1",
      "description": "...",
      "deadline": "2024-12-31T23:59:59Z",
      ...
    }
  ]
}
```

---

#### 2. Create Course Material
```
POST /api/course-materials
```
**Description:** สร้าง material ใหม่

**Authentication:** API Key + JWT (Teacher/Admin only)

**Request Body:**
```json
{
  "course_id": 1,
  "type": "code_exercise",
  "title": "Exercise 1",
  "description": "Exercise description",
  "week": 1,
  "deadline": "2024-12-31T23:59:59Z",
  "problem": "Write a function...",
  "is_published": true
}
```

---

#### 3. Get Course Material
```
GET /api/course-materials/:id
```
**Description:** ดึงข้อมูล material ตาม ID

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Material ID

---

#### 4. Update Course Material
```
PUT /api/course-materials/:id
```
**Description:** อัปเดตข้อมูล material

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Material ID

---

#### 5. Delete Course Material
```
DELETE /api/course-materials/:id
```
**Description:** ลบ material

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Material ID

---

#### 6. Upload Course Material File
```
POST /api/course-materials/upload
```
**Description:** อัปโหลดไฟล์สำหรับ material

**Authentication:** API Key + JWT (Teacher/Admin only)

**Request:** Multipart form data
- `file` (file, required) - ไฟล์ที่ต้องการอัปโหลด
- `course_id` (int, required) - Course ID
- `type` (string, required) - Material type

---

#### 7. Upload Problem Image
```
POST /api/course-materials/:id/images
```
**Description:** อัปโหลดรูปภาพสำหรับ problem

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Material ID

**Request:** Multipart form data
- `image` (file, required) - รูปภาพ

---

### Test Cases

#### 1. Get Test Cases
```
GET /api/course-materials/:id/test-cases
GET /api/test-cases/exercises/:exercise_id
```
**Description:** ดึงรายการ test cases ของ material/exercise

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Material ID
- `exercise_id` (int, required) - Exercise ID (สำหรับ endpoint ที่สอง)

---

#### 2. Add Test Case
```
POST /api/course-materials/:id/test-cases
POST /api/test-cases/exercises/:exercise_id
```
**Description:** เพิ่ม test case ใหม่

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Material ID
- `exercise_id` (int, required) - Exercise ID (สำหรับ endpoint ที่สอง)

**Request Body:**
```json
{
  "input": "test input",
  "expected_output": "expected output",
  "is_hidden": false,
  "order": 1
}
```

---

#### 3. Update Test Case
```
PUT /api/course-materials/test-cases/:test_case_id
PUT /api/test-cases/:id
```
**Description:** อัปเดต test case

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `test_case_id` (int, required) - Test Case ID
- `id` (int, required) - Test Case ID (สำหรับ endpoint ที่สอง)

---

#### 4. Delete Test Case
```
DELETE /api/course-materials/test-cases/:test_case_id
DELETE /api/test-cases/:id
```
**Description:** ลบ test case

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `test_case_id` (int, required) - Test Case ID
- `id` (int, required) - Test Case ID (สำหรับ endpoint ที่สอง)

---

### Submissions

#### 1. Submit Exercise
```
POST /api/course-materials/:id/submit
POST /api/submissions/exercises/:id
```
**Description:** ส่งคำตอบสำหรับ exercise

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Material ID หรือ Exercise ID

**Request Body:**
```json
{
  "code": "def solution():\n    return 42",
  "language": "python"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "status": "pending",
    "queue_job_id": 123,
    ...
  }
}
```

---

#### 2. Submit PDF Exercise
```
POST /api/course-materials/:id/submit-pdf
POST /api/materials/:material_id/submit
```
**Description:** ส่งไฟล์ PDF สำหรับ PDF exercise

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Material ID
- `material_id` (int, required) - Material ID (สำหรับ endpoint ที่สอง)

**Request:** Multipart form data
- `file` (file, required) - PDF file

---

#### 3. Get My Submission
```
GET /api/course-materials/:id/submissions/me
GET /api/materials/:material_id/submissions/me
```
**Description:** ดึง submission ของตัวเองสำหรับ material

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Material ID
- `material_id` (int, required) - Material ID (สำหรับ endpoint ที่สอง)

---

#### 4. List Exercise Submissions
```
GET /api/submissions/exercises/:id
```
**Description:** ดึงรายการ submissions ของ exercise

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Exercise ID

---

#### 5. Get Submission
```
GET /api/submissions/:id
```
**Description:** ดึงข้อมูล submission ตาม ID

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Submission ID

---

### Progress

#### 1. Get Self Progress
```
GET /api/students/progress
```
**Description:** ดึง progress ของตัวเอง

**Authentication:** API Key + JWT

**Query Parameters:**
- `course_id` (int, optional) - Filter by course

---

#### 2. Get Course Progress
```
GET /api/courses/:id/progress
```
**Description:** ดึง progress ของ course

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `id` (int, required) - Course ID

---

#### 3. Verify Progress
```
POST /api/progress/:id/verify
```
**Description:** ตรวจสอบและ verify progress

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `id` (int, required) - Progress ID

**Request Body:**
```json
{
  "status": "approved",
  "feedback": "Good work!"
}
```

---

#### 4. Get Verification Logs
```
GET /api/progress/:id/logs
```
**Description:** ดึง logs การ verify

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Progress ID

---

#### 5. Request Approval
```
POST /api/progress/:material_id/request-approval
```
**Description:** ขอ approval สำหรับ progress

**Authentication:** API Key + JWT

**Path Parameters:**
- `material_id` (int, required) - Material ID

---

### Announcements

#### 1. List Announcements
```
GET /api/announcements
```
**Description:** ดึงรายการ announcements

**Authentication:** API Key + JWT

**Query Parameters:**
- `course_id` (int, required) - Course ID
- `page` (int, optional) - Page number
- `limit` (int, optional) - Items per page

---

#### 2. Get Announcement
```
GET /api/announcements/:id
```
**Description:** ดึงข้อมูล announcement ตาม ID

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Announcement ID

---

#### 3. Create Announcement
```
POST /api/announcements
```
**Description:** สร้าง announcement ใหม่

**Authentication:** API Key + JWT (Teacher/Admin only)

**Request Body:**
```json
{
  "course_id": 1,
  "title": "Announcement Title",
  "content": "Announcement content",
  "is_pinned": false
}
```

---

#### 4. Update Announcement
```
PUT /api/announcements/:id
```
**Description:** อัปเดต announcement

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Announcement ID

---

#### 5. Delete Announcement
```
DELETE /api/announcements/:id
```
**Description:** ลบ announcement

**Authentication:** API Key + JWT (Teacher/Admin only)

**Path Parameters:**
- `id` (int, required) - Announcement ID

---

#### 6. Get Announcement Stats
```
GET /api/announcements/stats
```
**Description:** ดึงสถิติของ announcements

**Authentication:** API Key + JWT

**Query Parameters:**
- `course_id` (int, required) - Course ID

---

#### 7. Get Recent Announcements
```
GET /api/announcements/recent
```
**Description:** ดึง announcements ล่าสุด

**Authentication:** API Key + JWT

**Query Parameters:**
- `limit` (int, optional) - Number of announcements (default: 10)

---

### Course Scores

#### 1. Get Course Score
```
GET /api/course-scores/course
```
**Description:** ดึงคะแนนของ course

**Authentication:** API Key + JWT

**Query Parameters:**
- `course_id` (int, required) - Course ID

---

### Deadline Checker

#### 1. Check Material Deadline
```
GET /api/materials/check-deadline
```
**Description:** ตรวจสอบ deadline ของ material

**Authentication:** ไม่ต้องใช้

**Query Parameters:**
- `material_id` (int, required) - Material ID

---

#### 2. Get Available Materials
```
GET /api/materials/available
```
**Description:** ดึง materials ที่ยังเปิดให้ส่งได้

**Authentication:** API Key + JWT

**Query Parameters:**
- `course_id` (int, required) - Course ID

---

#### 3. Get Expired Materials
```
GET /api/materials/expired
```
**Description:** ดึง materials ที่หมดเวลาแล้ว

**Authentication:** API Key + JWT

**Query Parameters:**
- `course_id` (int, required) - Course ID

---

#### 4. Can Submit Exercise
```
GET /api/materials/can-submit
```
**Description:** ตรวจสอบว่าสามารถส่ง exercise ได้หรือไม่

**Authentication:** API Key + JWT

**Query Parameters:**
- `exercise_id` (int, required) - Exercise ID

---

#### 5. Get Materials by Deadline Status
```
GET /api/materials/by-deadline-status
```
**Description:** ดึง materials แยกตาม deadline status

**Authentication:** ไม่ต้องใช้

**Query Parameters:**
- `course_id` (int, required) - Course ID

---

#### 6. Get Upcoming Deadlines
```
GET /api/materials/upcoming-deadlines
```
**Description:** ดึง deadlines ที่กำลังจะมาถึง

**Authentication:** ไม่ต้องใช้

**Query Parameters:**
- `course_id` (int, required) - Course ID
- `hours` (int, optional) - Hours ahead (default: 24)

---

#### 7. Get Deadline Stats
```
GET /api/materials/deadline-stats
```
**Description:** ดึงสถิติของ deadlines

**Authentication:** ไม่ต้องใช้

**Query Parameters:**
- `course_id` (int, required) - Course ID

---

### Queue Management

#### 1. Get Queue Jobs
```
GET /api/queue/jobs
```
**Description:** ดึงรายการ queue jobs

**Authentication:** API Key + JWT

**Query Parameters:**
- `status` (string, optional) - Filter by status
- `type` (string, optional) - Filter by type
- `page` (int, optional) - Page number
- `limit` (int, optional) - Items per page

---

#### 2. Get Queue Job
```
GET /api/queue/jobs/:id
```
**Description:** ดึงข้อมูล queue job ตาม ID

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Job ID

---

#### 3. Cancel Queue Job
```
POST /api/queue/jobs/:id/cancel
```
**Description:** ยกเลิก queue job

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Job ID

---

#### 4. Process Queue Job
```
POST /api/queue/jobs/:id/process
```
**Description:** ประมวลผล queue job

**Authentication:** API Key + JWT (Admin only)

**Path Parameters:**
- `id` (int, required) - Job ID

---

#### 5. Claim Queue Job
```
POST /api/queue/jobs/:id/claim
```
**Description:** Claim queue job สำหรับ processing

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Job ID

---

#### 6. Complete Queue Job
```
POST /api/queue/jobs/:id/complete
```
**Description:** เสร็จสิ้น queue job

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Job ID

**Request Body:**
```json
{
  "result": "success",
  "output": "...",
  "error": null
}
```

---

#### 7. Retry Queue Job
```
POST /api/queue/jobs/:id/retry
```
**Description:** Retry queue job ที่ล้มเหลว

**Authentication:** API Key + JWT

**Path Parameters:**
- `id` (int, required) - Job ID

---

#### 8. Get Queue Stats
```
GET /api/queue/stats
```
**Description:** ดึงสถิติของ queue

**Authentication:** API Key + JWT

---

#### 9. Submit Code Review
```
POST /api/queue/review
```
**Description:** ส่ง code สำหรับ review

**Authentication:** API Key + JWT

**Request Body:**
```json
{
  "submission_id": 1,
  "code": "...",
  "language": "python"
}
```

---

### Drafts

#### 1. Save Draft
```
POST /api/drafts/exercises/:exercise_id
```
**Description:** บันทึก draft ของ exercise

**Authentication:** API Key + JWT

**Path Parameters:**
- `exercise_id` (int, required) - Exercise ID

**Request Body:**
```json
{
  "code": "def solution():\n    return 42"
}
```

---

#### 2. Get Draft
```
GET /api/drafts/exercises/:exercise_id
```
**Description:** ดึง draft ของ exercise

**Authentication:** API Key + JWT

**Path Parameters:**
- `exercise_id` (int, required) - Exercise ID

---

#### 3. Delete Draft
```
DELETE /api/drafts/exercises/:exercise_id
```
**Description:** ลบ draft

**Authentication:** API Key + JWT

**Path Parameters:**
- `exercise_id` (int, required) - Exercise ID

---

#### 4. Upload Python File (Draft)
```
POST /api/drafts/exercises/:exercise_id/upload
```
**Description:** อัปโหลดไฟล์ Python สำหรับ draft

**Authentication:** API Key + JWT

**Path Parameters:**
- `exercise_id` (int, required) - Exercise ID

**Request:** Multipart form data
- `file` (file, required) - Python file

---

#### 5. Get My Drafts
```
GET /api/drafts/my
```
**Description:** ดึง drafts ทั้งหมดของตัวเอง

**Authentication:** API Key + JWT

---

### PDF Exercise Submissions

#### 1. Get My PDF Submission
```
GET /api/materials/:material_id/submissions/me
```
**Description:** ดึง PDF submission ของตัวเอง

**Authentication:** API Key + JWT

**Path Parameters:**
- `material_id` (int, required) - Material ID

---

#### 2. Get PDF Submissions
```
GET /api/materials/:material_id/submissions
```
**Description:** ดึง PDF submissions ทั้งหมด (Teacher/TA only)

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `material_id` (int, required) - Material ID

---

#### 3. Get Course PDF Submissions
```
GET /api/courses/:course_id/pdf-submissions
```
**Description:** ดึง PDF submissions ทั้งหมดของ course

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `course_id` (int, required) - Course ID

---

#### 4. Get PDF Submission
```
GET /api/submissions/:submission_id
```
**Description:** ดึงข้อมูล PDF submission ตาม ID

**Authentication:** API Key + JWT

**Path Parameters:**
- `submission_id` (int, required) - Submission ID

---

#### 5. Approve PDF Submission
```
POST /api/submissions/:submission_id/approve
```
**Description:** อนุมัติ PDF submission

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `submission_id` (int, required) - Submission ID

**Request Body:**
```json
{
  "feedback": "Good work!",
  "score": 100
}
```

---

#### 6. Reject PDF Submission
```
POST /api/submissions/:submission_id/reject
```
**Description:** ปฏิเสธ PDF submission

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `submission_id` (int, required) - Submission ID

**Request Body:**
```json
{
  "feedback": "Needs improvement",
  "reason": "Incomplete"
}
```

---

#### 7. Download PDF Submission
```
GET /api/submissions/:submission_id/download
```
**Description:** ดาวน์โหลด PDF submission

**Authentication:** API Key + JWT (Teacher/TA/Admin only)

**Path Parameters:**
- `submission_id` (int, required) - Submission ID

---

#### 8. Download Feedback File
```
GET /api/submissions/:submission_id/feedback/download
```
**Description:** ดาวน์โหลดไฟล์ feedback

**Authentication:** API Key + JWT

**Path Parameters:**
- `submission_id` (int, required) - Submission ID

---

#### 9. Cancel PDF Submission
```
DELETE /api/submissions/:submission_id/cancel
```
**Description:** ยกเลิก PDF submission

**Authentication:** API Key + JWT

**Path Parameters:**
- `submission_id` (int, required) - Submission ID

---

### Playground (Go Backend)

#### 1. Run Code (Gateway)
```
POST /api/playground/run
```
**Description:** Gateway สำหรับรัน code (forward ไปยัง FastAPI)

**Authentication:** ไม่ต้องใช้

**Request Body:**
```json
{
  "code": "def solution():\n    return 42",
  "dataType": "singlylinkedlist"
}
```

---

#### 2. Playground Health Check
```
GET /api/playground/health
```
**Description:** ตรวจสอบสถานะของ playground service

**Authentication:** ไม่ต้องใช้

---

### System

#### 1. Health Check
```
GET /health
```
**Description:** ตรวจสอบสถานะของระบบ

**Authentication:** ไม่ต้องใช้

---

#### 2. Secure Health Check
```
GET /health/secure
```
**Description:** ตรวจสอบสถานะของระบบ (ต้องใช้ API Key)

**Authentication:** API Key

---

#### 3. API Information
```
GET /
```
**Description:** ดึงข้อมูล API และ endpoints ทั้งหมด

**Authentication:** API Key

---

#### 4. Test Public Endpoint
```
GET /test-public
```
**Description:** Test endpoint (public)

**Authentication:** ไม่ต้องใช้

---

## FastAPI Backend API

### Playground

#### 1. Run Code
```
POST /api/playground/run
```
**Description:** Execute code และส่งคืน execution steps

**Authentication:** API Key (required)

**Request Body:**
```json
{
  "code": "class Node:\n    def __init__(self, data):\n        self.data = data\n        self.next = None",
  "dataType": "singlylinkedlist"
}
```

**Supported Data Types:**
- `singlylinkedlist`
- `doublylinkedlist`
- `stack`
- `binarysearchtree`
- `undirectedgraph`
- `directedgraph`
- `queue`

**Response:**
```json
{
  "executionId": "uuid",
  "status": "success",
  "steps": [
    {
      "step": 1,
      "action": "create",
      "data": {...}
    }
  ],
  "result": "..."
}
```

---

### System

#### 1. Root Endpoint
```
GET /
```
**Description:** ข้อมูล API

**Authentication:** ไม่ต้องใช้

---

#### 2. Health Check
```
GET /health
```
**Description:** ตรวจสอบสถานะของระบบ

**Authentication:** ไม่ต้องใช้

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "0.0.5-alpha",
  "service": "DSView Backend API",
  "uptime": "running"
}
```

---

## Error Codes

### HTTP Status Codes

- `200 OK` - Request สำเร็จ
- `201 Created` - สร้าง resource สำเร็จ
- `400 Bad Request` - Request ไม่ถูกต้อง
- `401 Unauthorized` - ไม่มี authentication
- `403 Forbidden` - ไม่มีสิทธิ์เข้าถึง
- `404 Not Found` - ไม่พบ resource
- `409 Conflict` - มี conflict (เช่น duplicate)
- `422 Unprocessable Entity` - Validation error
- `500 Internal Server Error` - Server error

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": {}
  }
}
```

---

## Roles & Permissions

### Student
- ดู courses ที่ลงทะเบียน
- ดู materials ที่ published
- ส่ง submissions
- ดู progress ของตัวเอง

### TA (Teaching Assistant)
- สิทธิ์ของ Student
- ดู enrollments
- ดู submissions ทั้งหมด
- Verify progress
- ดู reports

### Teacher
- สิทธิ์ของ TA
- สร้าง/แก้ไข/ลบ courses
- สร้าง/แก้ไข/ลบ materials
- สร้าง/แก้ไข/ลบ test cases
- สร้าง/แก้ไข/ลบ announcements
- ดู teacher reports

### Admin
- สิทธิ์ทั้งหมด
- จัดการ users
- จัดการระบบ

---

## Rate Limiting

- **General API**: 100 requests per minute
- **Auth endpoints**: Stricter rate limiting (configurable)
- **Playground**: Configurable per minute

---

## Notes

1. **Base URL**: ตรวจสอบ config สำหรับ base URL ที่ถูกต้อง
2. **Authentication**: ส่วนใหญ่ endpoints ต้องการ API Key และ/หรือ JWT Token
3. **Enrollment Validation**: หลาย endpoints ตรวจสอบ enrollment ก่อนเข้าถึง
4. **File Uploads**: ใช้ multipart/form-data สำหรับ file uploads
5. **Pagination**: หลาย endpoints รองรับ pagination ผ่าน `page` และ `limit` parameters

---

## Swagger Documentation

- **Go Backend Swagger UI**: `/docs/`
- **Go Backend OpenAPI JSON**: `/docs/doc.json`
- **FastAPI Swagger UI**: `/docs`
- **FastAPI ReDoc**: `/redoc`

---

*Last Updated: 2024*












