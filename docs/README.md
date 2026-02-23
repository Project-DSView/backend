# DSView Backend API Documentation

ยินดีต้อนรับสู่เอกสาร API ของ DSView Backend

## เอกสารที่มี

1. **[API_DOCUMENTATION.md](./API_DOCUMENTATION.md)** - เอกสาร API ฉบับเต็ม
   - รายละเอียด endpoints ทั้งหมด
   - Request/Response examples
   - Authentication guide
   - Error codes

## โครงสร้าง Backend

ระบบ DSView Backend ประกอบด้วย 2 ส่วนหลัก:

### 1. Go Backend (`backend/go/`)
- API หลักสำหรับจัดการระบบ
- Authentication & Authorization
- Course Management
- Material Management
- Submissions & Progress
- Queue Management

**Base URL**: `http://localhost:8080` (หรือตาม config)

**Swagger UI**: `/docs/`

### 2. FastAPI Backend (`backend/fastapi/`)
- Code Execution Service
- Playground API
- Step-by-step visualization

**Base URL**: `http://localhost:8000` (หรือตาม config)

**Swagger UI**: `/docs`

## การเริ่มต้นใช้งาน

### Authentication

1. **OAuth Login (Google)**
   ```
   GET /api/auth/google
   ```

2. **Get JWT Token**
   - หลังจาก login สำเร็จ จะได้รับ JWT token
   - ใช้ token นี้ใน header: `Authorization: Bearer <token>`

3. **API Key**
   - ต้องใช้ API Key สำหรับ endpoints ที่ต้องการความปลอดภัย
   - Header: `X-API-Key: <your-api-key>`

### ตัวอย่างการใช้งาน

#### 1. Login และ Get Profile
```bash
# 1. Login ผ่าน Google OAuth
curl -X GET "http://localhost:8080/api/auth/google"

# 2. ใช้ JWT token ที่ได้
curl -X GET "http://localhost:8080/api/profile" \
  -H "Authorization: Bearer <jwt-token>" \
  -H "X-API-Key: <api-key>"
```

#### 2. Get Courses
```bash
curl -X GET "http://localhost:8080/api/courses" \
  -H "Authorization: Bearer <jwt-token>" \
  -H "X-API-Key: <api-key>"
```

#### 3. Submit Exercise
```bash
curl -X POST "http://localhost:8080/api/course-materials/1/submit" \
  -H "Authorization: Bearer <jwt-token>" \
  -H "X-API-Key: <api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "def solution():\n    return 42",
    "language": "python"
  }'
```

#### 4. Run Code in Playground
```bash
curl -X POST "http://localhost:8000/api/playground/run" \
  -H "X-API-Key: <api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "class Node:\n    def __init__(self, data):\n        self.data = data",
    "dataType": "singlylinkedlist"
  }'
```

## Endpoints หลัก

### Authentication
- `GET /api/auth/google` - Google OAuth login
- `GET /api/auth/google/callback` - OAuth callback
- `POST /api/auth/logout` - Logout
- `POST /api/auth/refresh` - Refresh token
- `GET /api/profile` - Get user profile

### Courses
- `GET /api/courses` - List courses
- `POST /api/courses` - Create course
- `GET /api/courses/:id` - Get course
- `PUT /api/courses/:id` - Update course
- `DELETE /api/courses/:id` - Delete course

### Course Materials
- `GET /api/course-materials` - List materials
- `POST /api/course-materials` - Create material
- `GET /api/course-materials/:id` - Get material
- `PUT /api/course-materials/:id` - Update material
- `DELETE /api/course-materials/:id` - Delete material

### Submissions
- `POST /api/course-materials/:id/submit` - Submit exercise
- `GET /api/submissions/:id` - Get submission
- `GET /api/course-materials/:id/submissions/me` - Get my submission

### Playground (FastAPI)
- `POST /api/playground/run` - Run code
- `GET /health` - Health check

## Roles & Permissions

| Role | Permissions |
|------|-------------|
| **Student** | View courses, materials, submit exercises, view own progress |
| **TA** | Student permissions + view all submissions, verify progress, view reports |
| **Teacher** | TA permissions + create/edit/delete courses, materials, test cases, announcements |
| **Admin** | All permissions + user management, system administration |

## Error Handling

### Standard Error Response
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

### Common Error Codes
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Validation Error
- `500` - Internal Server Error

## Rate Limiting

- **General API**: 100 requests/minute
- **Auth endpoints**: Stricter limits
- **Playground**: Configurable

## Additional Resources

- **Swagger UI (Go)**: `http://localhost:8080/docs/`
- **Swagger UI (FastAPI)**: `http://localhost:8000/docs`
- **OpenAPI JSON (Go)**: `http://localhost:8080/docs/doc.json`

## Support

สำหรับคำถามหรือปัญหา กรุณาติดต่อทีมพัฒนา

---

*Last Updated: 2024*












