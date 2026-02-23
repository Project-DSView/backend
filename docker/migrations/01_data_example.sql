-- ข้อมูลตัวอย่างสำหรับระบบการเรียนรู้โครงสร้างข้อมูล
-- Generated based on database schema and Go models

-- =============================================
-- 1. USERS (ผู้ใช้ในระบบ)
-- =============================================

-- ครูผู้สอน
INSERT INTO users (user_id, first_name, last_name, email, is_teacher, profile_img, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'สมชาย', 'ใจดี', 'somchai@kmitl.ac.th', true, 'https://example.com/profiles/somchai.jpg', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'สมหญิง', 'สอนดี', 'somying@kmitl.ac.th', true, 'https://example.com/profiles/somying.jpg', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'อาจารย์', 'ผู้เชี่ยวชาญ', 'expert@kmitl.ac.th', true, NULL, NOW(), NOW());

-- นักเรียน
INSERT INTO users (user_id, first_name, last_name, email, is_teacher, profile_img, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440010', 'นักเรียน', 'คนที่หนึ่ง', 'student1@kmitl.ac.th', false, NULL, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440011', 'นักเรียน', 'คนที่สอง', 'student2@kmitl.ac.th', false, NULL, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440012', 'นักเรียน', 'คนที่สาม', 'student3@kmitl.ac.th', false, NULL, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440013', 'นักเรียน', 'คนที่สี่', 'student4@kmitl.ac.th', false, NULL, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440014', 'นักเรียน', 'คนที่ห้า', 'student5@kmitl.ac.th', false, NULL, NOW(), NOW());

-- ผู้ช่วยสอน (TA)
INSERT INTO users (user_id, first_name, last_name, email, is_teacher, profile_img, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440020', 'ผู้ช่วยสอน', 'คนที่หนึ่ง', 'ta1@kmitl.ac.th', false, NULL, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440021', 'ผู้ช่วยสอน', 'คนที่สอง', 'ta2@kmitl.ac.th', false, NULL, NOW(), NOW());

-- =============================================
-- 2. COURSES (คอร์สเรียน)
-- =============================================

INSERT INTO courses (course_id, name, description, image_url, created_by, enroll_key, status, created_at, updated_at) VALUES
('650e8400-e29b-41d4-a716-446655440001', 'โครงสร้างข้อมูลและอัลกอริทึม', 'เรียนรู้พื้นฐานโครงสร้างข้อมูลและอัลกอริทึมสำหรับการเขียนโปรแกรม', 'https://example.com/courses/ds-algo.jpg', '550e8400-e29b-41d4-a716-446655440001', 'DS2024A', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440002', 'การเขียนโปรแกรมขั้นสูง', 'เรียนรู้เทคนิคการเขียนโปรแกรมขั้นสูงและ Design Patterns', 'https://example.com/courses/advanced-prog.jpg', '550e8400-e29b-41d4-a716-446655440002', 'ADV2024B', 'active', NOW(), NOW()),
('650e8400-e29b-41d4-a716-446655440003', 'ฐานข้อมูลและระบบจัดการข้อมูล', 'เรียนรู้การออกแบบและจัดการฐานข้อมูล', 'https://example.com/courses/database.jpg', '550e8400-e29b-41d4-a716-446655440003', 'DB2024C', 'archived', NOW(), NOW());

-- =============================================
-- 3. ENROLLMENTS (การลงทะเบียนเรียน)
-- =============================================

-- ลงทะเบียนในคอร์สโครงสร้างข้อมูล
INSERT INTO enrollments (enrollment_id, course_id, user_id, role, enrolled_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'teacher', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440010', 'student', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440011', 'student', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440012', 'student', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440013', 'student', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440014', 'student', NOW(), NOW());

-- ลงทะเบียนในคอร์สการเขียนโปรแกรมขั้นสูง
INSERT INTO enrollments (enrollment_id, course_id, user_id, role, enrolled_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440007', '650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', 'teacher', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440008', '650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440010', 'student', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440009', '650e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440011', 'student', NOW(), NOW());

-- ลงทะเบียนผู้ช่วยสอน (TA) ในคอร์สโครงสร้างข้อมูล
INSERT INTO enrollments (enrollment_id, course_id, user_id, role, enrolled_at, updated_at) VALUES
('750e8400-e29b-41d4-a716-446655440010', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440020', 'ta', NOW(), NOW()),
('750e8400-e29b-41d4-a716-446655440011', '650e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440021', 'ta', NOW(), NOW());

-- =============================================
-- 4. COURSE_WEEKS (สัปดาห์การเรียน)
-- =============================================

-- สัปดาห์สำหรับคอร์สโครงสร้างข้อมูล
INSERT INTO course_weeks (course_week_id, course_id, week_number, title, description, created_by, created_at, updated_at) VALUES
('850e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001', 1, 'แนะนำโครงสร้างข้อมูล', 'เรียนรู้พื้นฐานและความสำคัญของโครงสร้างข้อมูล', '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440001', 2, 'Array และ Linked List', 'เรียนรู้การใช้งาน Array และ Linked List', '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440001', 3, 'Stack และ Queue', 'เรียนรู้โครงสร้างข้อมูล Stack และ Queue', '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440001', 4, 'Tree และ Binary Tree', 'เรียนรู้โครงสร้างข้อมูลแบบ Tree', '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('850e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440001', 5, 'Graph และ Graph Algorithms', 'เรียนรู้โครงสร้างข้อมูล Graph และอัลกอริทึม', '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- =============================================
-- 5. MATERIALS (เนื้อหาและแบบฝึกหัด - แยกตามตาราง)
-- =============================================

-- เอกสารประกอบการเรียน (Documents)
INSERT INTO documents (material_id, course_id, title, description, week, file_url, file_name, file_size, mime_type, is_public, created_by, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001', 'บทที่ 1: แนะนำโครงสร้างข้อมูล', 'เอกสารประกอบการเรียนบทที่ 1', 1, 'https://example.com/files/chapter1.pdf', 'chapter1.pdf', 2048576, 'application/pdf', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440001', 'บทที่ 2: Array และ Linked List', 'เอกสารประกอบการเรียนบทที่ 2', 2, 'https://example.com/files/chapter2.pdf', 'chapter2.pdf', 1536000, 'application/pdf', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440001', 'บทที่ 3: Stack และ Queue', 'เอกสารประกอบการเรียนบทที่ 3', 3, 'https://example.com/files/chapter3.pdf', 'chapter3.pdf', 1873408, 'application/pdf', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- วิดีโอการสอน (Videos)
INSERT INTO videos (material_id, course_id, title, description, week, video_url, is_public, created_by, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440001', 'วิดีโอสอน: Array และ Linked List', 'วิดีโออธิบายการใช้งาน Array และ Linked List', 2, 'https://www.youtube.com/watch?v=dQw4w9WgXcQ', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440001', 'วิดีโอสอน: Stack และ Queue', 'วิดีโออธิบายการใช้งาน Stack และ Queue', 3, 'https://www.youtube.com/watch?v=dQw4w9WgXcQ', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- แบบฝึกหัดโค้ด (Code Exercises)
INSERT INTO code_exercises (material_id, course_id, title, description, week, total_points, deadline, is_graded, problem_statement, example_inputs, example_outputs, constraints, hints, is_public, created_by, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', 'แบบฝึกหัด: Array Operations', 'เขียนฟังก์ชันสำหรับการทำงานกับ Array', 2, 100, '2024-12-31T23:59:59Z', true, 
'เขียนฟังก์ชัน `findMax` ที่รับ array ของตัวเลขและคืนค่าตัวเลขที่มากที่สุด

**ตัวอย่าง:**
- Input: [1, 5, 3, 9, 2]
- Output: 9

**ข้อกำหนด:**
- ฟังก์ชันต้องรับ parameter เป็น array ของ int
- คืนค่าเป็น int
- ใช้เวลา O(n)',
'["[1, 5, 3, 9, 2]", "[10, 20, 30]", "[-1, -5, -3]"]',
'["9", "30", "-1"]',
'1 ≤ n ≤ 1000
-1000 ≤ arr[i] ≤ 1000',
'ลองใช้ตัวแปรเก็บค่าสูงสุดและเปรียบเทียบทีละตัว',
true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),

('950e8400-e29b-41d4-a716-446655440007', '650e8400-e29b-41d4-a716-446655440001', 'แบบฝึกหัด: Stack Implementation', 'สร้าง Stack class พร้อมฟังก์ชันพื้นฐาน', 3, 150, '2024-12-31T23:59:59Z', true,
'สร้าง Stack class ที่มีฟังก์ชัน push, pop, peek, และ isEmpty

**ตัวอย่าง:**
```python
stack = Stack()
stack.push(1)
stack.push(2)
print(stack.peek())  # 2
print(stack.pop())   # 2
print(stack.isEmpty())  # False
```',
'["push(1), push(2), peek()", "push(5), pop(), isEmpty()", "isEmpty()"]',
'["2", "False", "True"]',
'1 ≤ operations ≤ 1000
-1000 ≤ value ≤ 1000',
'ใช้ list หรือ array ในการเก็บข้อมูล',
true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),

('950e8400-e29b-41d4-a716-446655440008', '650e8400-e29b-41d4-a716-446655440001', 'แบบฝึกหัด: Binary Tree Traversal', 'เขียนฟังก์ชันสำหรับการ traverse Binary Tree', 4, 200, '2024-12-31T23:59:59Z', true,
'เขียนฟังก์ชัน `inorderTraversal` ที่ทำ inorder traversal ของ binary tree

**ตัวอย่าง:**
```
    1
   / \
  2   3
 /
4
```
Output: [4, 2, 1, 3]',
'["[1,2,3,4,null,null,null]", "[1,null,2,3]", "[]"]',
'["[4,2,1,3]", "[1,3,2]", "[]"]',
'0 ≤ nodes ≤ 100
-100 ≤ node.val ≤ 100',
'ใช้ recursive approach หรือ iterative approach',
true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),

('950e8400-e29b-41d4-a716-446655440011', '650e8400-e29b-41d4-a716-446655440001', 'แบบฝึกหัดฝึกฝน: Basic Sorting', 'ฝึกเขียนอัลกอริทึมการเรียงลำดับพื้นฐาน', 3, 50, '2025-12-15T23:59:59Z', false, 
'เขียนฟังก์ชัน `bubbleSort` ที่รับ array และเรียงลำดับแบบ bubble sort

**ตัวอย่าง:**
- Input: [3, 1, 4, 1, 5]
- Output: [1, 1, 3, 4, 5]

**ข้อจำกัด:**
- ใช้ bubble sort algorithm เท่านั้น
- Time complexity: O(n²)', 
'["[3, 1, 4, 1, 5]", "[5, 4, 3, 2, 1]"]', 
'["[1, 1, 3, 4, 5]", "[1, 2, 3, 4, 5]"]', 
'Array size 1-1000 elements', 
'ลองเขียน step by step ตาม bubble sort algorithm',
true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- แบบฝึกหัด PDF (PDF Exercises)
INSERT INTO pdf_exercises (material_id, course_id, title, description, week, total_points, deadline, is_graded, file_url, file_name, file_size, mime_type, is_public, created_by, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440009', '650e8400-e29b-41d4-a716-446655440001', 'แบบฝึกหัด: การวิเคราะห์อัลกอริทึม', 'แบบฝึกหัดการวิเคราะห์ Big O และอัลกอริทึม', 4, 100, '2025-10-31T23:59:59Z', true, 'https://example.com/files/algorithm_analysis.pdf', 'algorithm_analysis.pdf', 1024000, 'application/pdf', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440010', '650e8400-e29b-41d4-a716-446655440001', 'แบบฝึกหัด: Graph Algorithms', 'แบบฝึกหัดการเขียนอัลกอริทึมสำหรับ Graph', 3, 150, '2025-12-31T23:59:59Z', true, 'https://example.com/files/graph_algorithms.pdf', 'graph_algorithms.pdf', 1536000, 'application/pdf', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- ประกาศ (Announcements)
INSERT INTO announcements (material_id, course_id, title, description, week, content, is_public, created_by, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440012', '650e8400-e29b-41d4-a716-446655440001', 'ยินดีต้อนรับเข้าสู่คอร์สโครงสร้างข้อมูล', 'ยินดีต้อนรับทุกท่านเข้าสู่คอร์สเรียนโครงสร้างข้อมูลและอัลกอริทึม', 1, 'ยินดีต้อนรับทุกท่านเข้าสู่คอร์สเรียนโครงสร้างข้อมูลและอัลกอริทึม ในคอร์สนี้เราจะเรียนรู้พื้นฐานที่สำคัญสำหรับการเขียนโปรแกรม', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440013', '650e8400-e29b-41d4-a716-446655440001', 'กำหนดส่งแบบฝึกหัด', 'กำหนดส่งแบบฝึกหัด', 2, 'แบบฝึกหัดทั้งหมดต้องส่งภายในวันที่ 31 ธันวาคม 2567 เวลา 23:59 น. กรุณาตรวจสอบกำหนดส่งให้ดี', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440014', '650e8400-e29b-41d4-a716-446655440001', 'การสอบกลางภาค', 'การสอบกลางภาค', 3, 'การสอบกลางภาคจะจัดในวันที่ 15 มกราคม 2568 เวลา 09:00-12:00 น. ณ ห้อง 101', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- =============================================
-- 5.1. COURSE_MATERIALS (ตารางกลาง - References)
-- =============================================

-- References สำหรับ Documents
INSERT INTO course_materials (material_id, course_id, type, week, reference_id, reference_type, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001', 'document', 1, '950e8400-e29b-41d4-a716-446655440001', 'document', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440001', 'document', 2, '950e8400-e29b-41d4-a716-446655440002', 'document', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440003', '650e8400-e29b-41d4-a716-446655440001', 'document', 3, '950e8400-e29b-41d4-a716-446655440003', 'document', NOW(), NOW());

-- References สำหรับ Videos
INSERT INTO course_materials (material_id, course_id, type, week, reference_id, reference_type, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440001', 'video', 2, '950e8400-e29b-41d4-a716-446655440004', 'video', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440001', 'video', 3, '950e8400-e29b-41d4-a716-446655440005', 'video', NOW(), NOW());

-- References สำหรับ Code Exercises
INSERT INTO course_materials (material_id, course_id, type, week, reference_id, reference_type, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', 'code_exercise', 2, '950e8400-e29b-41d4-a716-446655440006', 'code_exercise', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440007', '650e8400-e29b-41d4-a716-446655440001', 'code_exercise', 3, '950e8400-e29b-41d4-a716-446655440007', 'code_exercise', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440008', '650e8400-e29b-41d4-a716-446655440001', 'code_exercise', 4, '950e8400-e29b-41d4-a716-446655440008', 'code_exercise', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440011', '650e8400-e29b-41d4-a716-446655440001', 'code_exercise', 3, '950e8400-e29b-41d4-a716-446655440011', 'code_exercise', NOW(), NOW());

-- References สำหรับ PDF Exercises
INSERT INTO course_materials (material_id, course_id, type, week, reference_id, reference_type, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440009', '650e8400-e29b-41d4-a716-446655440001', 'pdf_exercise', 4, '950e8400-e29b-41d4-a716-446655440009', 'pdf_exercise', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440010', '650e8400-e29b-41d4-a716-446655440001', 'pdf_exercise', 3, '950e8400-e29b-41d4-a716-446655440010', 'pdf_exercise', NOW(), NOW());

-- References สำหรับ Announcements
INSERT INTO course_materials (material_id, course_id, type, week, reference_id, reference_type, created_at, updated_at) VALUES
('950e8400-e29b-41d4-a716-446655440012', '650e8400-e29b-41d4-a716-446655440001', 'announcement', 1, '950e8400-e29b-41d4-a716-446655440012', 'announcement', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440013', '650e8400-e29b-41d4-a716-446655440001', 'announcement', 2, '950e8400-e29b-41d4-a716-446655440013', 'announcement', NOW(), NOW()),
('950e8400-e29b-41d4-a716-446655440014', '650e8400-e29b-41d4-a716-446655440001', 'announcement', 3, '950e8400-e29b-41d4-a716-446655440014', 'announcement', NOW(), NOW());

-- =============================================
-- 6. TEST_CASES (กรณีทดสอบ)
-- =============================================

-- Test cases สำหรับแบบฝึกหัด Array Operations
INSERT INTO test_cases (test_case_id, material_id, input_data, expected_output, is_public, display_name, created_at, updated_at) VALUES
('a50e8400-e29b-41d4-a716-446655440001', '950e8400-e29b-41d4-a716-446655440006', '{"arr": [1, 5, 3, 9, 2]}', '{"result": 9}', false, 'Test Case 1: Basic array', NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440002', '950e8400-e29b-41d4-a716-446655440006', '{"arr": [10, 20, 30]}', '{"result": 30}', false, 'Test Case 2: Ascending array', NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440003', '950e8400-e29b-41d4-a716-446655440006', '{"arr": [-1, -5, -3]}', '{"result": -1}', false, 'Test Case 3: Negative numbers', NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440004', '950e8400-e29b-41d4-a716-446655440006', '{"arr": [42]}', '{"result": 42}', true, 'Test Case 4: Single element', NOW(), NOW());

-- Test cases สำหรับแบบฝึกหัด Stack Implementation
INSERT INTO test_cases (test_case_id, material_id, input_data, expected_output, is_public, display_name, created_at, updated_at) VALUES
('a50e8400-e29b-41d4-a716-446655440005', '950e8400-e29b-41d4-a716-446655440007', '{"operations": ["push(1)", "push(2)", "peek()"]}', '{"result": 2}', false, 'Test Case 1: Basic stack operations', NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440006', '950e8400-e29b-41d4-a716-446655440007', '{"operations": ["push(5)", "pop()", "isEmpty()"]}', '{"result": true}', false, 'Test Case 2: Pop and check empty', NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440007', '950e8400-e29b-41d4-a716-446655440007', '{"operations": ["isEmpty()"]}', '{"result": true}', true, 'Test Case 3: Empty stack', NOW(), NOW());

-- Test cases สำหรับแบบฝึกหัด Binary Tree Traversal
INSERT INTO test_cases (test_case_id, material_id, input_data, expected_output, is_public, display_name, created_at, updated_at) VALUES
('a50e8400-e29b-41d4-a716-446655440008', '950e8400-e29b-41d4-a716-446655440008', '{"tree": [1,2,3,4,null,null,null]}', '{"result": [4,2,1,3]}', false, 'Test Case 1: Basic tree', NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440009', '950e8400-e29b-41d4-a716-446655440008', '{"tree": [1,null,2,3]}', '{"result": [1,3,2]}', false, 'Test Case 2: Right skewed tree', NOW(), NOW()),
('a50e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440008', '{"tree": []}', '{"result": []}', true, 'Test Case 3: Empty tree', NOW(), NOW());

-- =============================================
-- 7. SUBMISSIONS (การส่งงาน)
-- =============================================

-- การส่งงานแบบฝึกหัดโค้ด
INSERT INTO submissions (submission_id, user_id, material_id, code, passed_count, failed_count, total_score, status, is_late_submission, submitted_at) VALUES
('b50e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440006', 
'def findMax(arr):
    if not arr:
        return None
    max_val = arr[0]
    for num in arr:
        if num > max_val:
            max_val = num
    return max_val', 4, 0, 100, 'completed', false, NOW()),

('b50e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440011', '950e8400-e29b-41d4-a716-446655440006', 
'def findMax(arr):
    return max(arr) if arr else None', 4, 0, 100, 'completed', false, NOW()),

('b50e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440012', '950e8400-e29b-41d4-a716-446655440006', 
'def findMax(arr):
    # Wrong implementation
    return arr[0] if arr else None', 1, 3, 25, 'completed', false, NOW()),

('b50e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440007', 
'class Stack:
    def __init__(self):
        self.items = []
    
    def push(self, item):
        self.items.append(item)
    
    def pop(self):
        return self.items.pop() if self.items else None
    
    def peek(self):
        return self.items[-1] if self.items else None
    
    def isEmpty(self):
        return len(self.items) == 0', 3, 0, 150, 'completed', false, NOW());

-- การส่งงานแบบฝึกหัด PDF
INSERT INTO submissions (submission_id, user_id, material_id, file_url, file_name, file_size, mime_type, status, is_late_submission, feedback, graded_at, graded_by, submitted_at) VALUES
('b50e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440009', 
'https://example.com/submissions/analysis_homework.pdf', 'analysis_homework.pdf', 1024000, 'application/pdf', 'completed', false, 'งานดีมาก ครบถ้วนตามที่กำหนด วิเคราะห์ได้ถูกต้อง', NOW(), '550e8400-e29b-41d4-a716-446655440001', NOW()),
('b50e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440011', '950e8400-e29b-41d4-a716-446655440009', 
'https://example.com/submissions/analysis_homework_v2.pdf', 'analysis_homework_v2.pdf', 1152000, 'application/pdf', 'pending', false, NULL, NULL, NULL, NOW()),
('b50e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440012', '950e8400-e29b-41d4-a716-446655440010', 
'https://example.com/submissions/graph_algorithms.pdf', 'graph_algorithms.pdf', 2048000, 'application/pdf', 'pending', false, NULL, NULL, NULL, NOW());

-- การส่งงานแบบฝึกหัดฝึกฝน (Practice exercise) - ส่งช้า
INSERT INTO submissions (submission_id, user_id, material_id, code, passed_count, failed_count, total_score, status, is_late_submission, submitted_at) VALUES
('b50e8400-e29b-41d4-a716-446655440008', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440011', 
'def bubbleSort(arr):
    n = len(arr)
    for i in range(n):
        for j in range(0, n - i - 1):
            if arr[j] > arr[j + 1]:
                arr[j], arr[j + 1] = arr[j + 1], arr[j]
    return arr', 2, 0, 50, 'completed', true, NOW());

-- =============================================
-- 8. SUBMISSION_RESULTS (ผลการทดสอบ)
-- =============================================

-- ผลการทดสอบสำหรับการส่งงาน Array Operations
INSERT INTO submission_results (result_id, submission_id, test_case_id, status, actual_output, created_at) VALUES
('c50e8400-e29b-41d4-a716-446655440001', 'b50e8400-e29b-41d4-a716-446655440001', 'a50e8400-e29b-41d4-a716-446655440001', 'passed', '{"result": 9}', NOW()),
('c50e8400-e29b-41d4-a716-446655440002', 'b50e8400-e29b-41d4-a716-446655440001', 'a50e8400-e29b-41d4-a716-446655440002', 'passed', '{"result": 30}', NOW()),
('c50e8400-e29b-41d4-a716-446655440003', 'b50e8400-e29b-41d4-a716-446655440001', 'a50e8400-e29b-41d4-a716-446655440003', 'passed', '{"result": -1}', NOW()),
('c50e8400-e29b-41d4-a716-446655440004', 'b50e8400-e29b-41d4-a716-446655440001', 'a50e8400-e29b-41d4-a716-446655440004', 'passed', '{"result": 42}', NOW()),

('c50e8400-e29b-41d4-a716-446655440005', 'b50e8400-e29b-41d4-a716-446655440002', 'a50e8400-e29b-41d4-a716-446655440001', 'passed', '{"result": 9}', NOW()),
('c50e8400-e29b-41d4-a716-446655440006', 'b50e8400-e29b-41d4-a716-446655440002', 'a50e8400-e29b-41d4-a716-446655440002', 'passed', '{"result": 30}', NOW()),
('c50e8400-e29b-41d4-a716-446655440007', 'b50e8400-e29b-41d4-a716-446655440002', 'a50e8400-e29b-41d4-a716-446655440003', 'passed', '{"result": -1}', NOW()),
('c50e8400-e29b-41d4-a716-446655440008', 'b50e8400-e29b-41d4-a716-446655440002', 'a50e8400-e29b-41d4-a716-446655440004', 'passed', '{"result": 42}', NOW()),

('c50e8400-e29b-41d4-a716-446655440009', 'b50e8400-e29b-41d4-a716-446655440003', 'a50e8400-e29b-41d4-a716-446655440001', 'failed', '{"result": 1}', NOW()),
('c50e8400-e29b-41d4-a716-446655440010', 'b50e8400-e29b-41d4-a716-446655440003', 'a50e8400-e29b-41d4-a716-446655440002', 'failed', '{"result": 10}', NOW()),
('c50e8400-e29b-41d4-a716-446655440011', 'b50e8400-e29b-41d4-a716-446655440003', 'a50e8400-e29b-41d4-a716-446655440003', 'failed', '{"result": -1}', NOW()),
('c50e8400-e29b-41d4-a716-446655440012', 'b50e8400-e29b-41d4-a716-446655440003', 'a50e8400-e29b-41d4-a716-446655440004', 'passed', '{"result": 42}', NOW());

-- =============================================
-- 9. STUDENT_PROGRESS (ความคืบหน้าของนักเรียน)
-- =============================================

INSERT INTO student_progress (progress_id, user_id, material_id, status, score, seat_number, last_submitted_at, created_at, updated_at) VALUES
('d50e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440006', 'completed', 100, 'A001', NOW(), NOW(), NOW()),
('d50e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440011', '950e8400-e29b-41d4-a716-446655440006', 'completed', 100, 'A002', NOW(), NOW(), NOW()),
('d50e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440012', '950e8400-e29b-41d4-a716-446655440006', 'completed', 25, 'A003', NOW(), NOW(), NOW()),
('d50e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440007', 'completed', 150, 'A001', NOW(), NOW(), NOW()),
('d50e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440009', 'waiting_review', 0, 'A001', NOW(), NOW(), NOW()),
('d50e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440011', '950e8400-e29b-41d4-a716-446655440009', 'waiting_review', 0, 'A002', NOW(), NOW(), NOW()),
('d50e8400-e29b-41d4-a716-446655440007', '550e8400-e29b-41d4-a716-446655440012', '950e8400-e29b-41d4-a716-446655440010', 'waiting_review', 0, 'A003', NOW(), NOW(), NOW());

-- =============================================
-- 10. VERIFICATION_LOGS (บันทึกการตรวจสอบ)
-- =============================================

INSERT INTO verification_logs (log_id, progress_id, verified_by, status, comment, verified_at) VALUES
('e50e8400-e29b-41d4-a716-446655440001', 'd50e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440001', 'approved', 'งานดีมาก ครบถ้วนตามที่กำหนด', NOW()),
('e50e8400-e29b-41d4-a716-446655440002', 'd50e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440001', 'rejected', 'ยังไม่ครบตามที่กำหนด กรุณาแก้ไขและส่งใหม่', NOW());

-- =============================================
-- 11. ANNOUNCEMENTS (ประกาศ)
-- =============================================

-- Additional announcements (using material_id from MaterialBase)
INSERT INTO announcements (material_id, course_id, title, description, week, content, is_public, created_by, created_at, updated_at) VALUES
('f50e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440001', 'ประกาศเพิ่มเติม 1', 'ประกาศเพิ่มเติมสำหรับคอร์ส', 1, 'เนื้อหาประกาศเพิ่มเติม 1', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('f50e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440001', 'ประกาศเพิ่มเติม 2', 'ประกาศเพิ่มเติมสำหรับคอร์ส', 2, 'เนื้อหาประกาศเพิ่มเติม 2', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW()),
('f50e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', 'ประกาศเพิ่มเติม 3', 'ประกาศเพิ่มเติมสำหรับคอร์ส', 3, 'เนื้อหาประกาศเพิ่มเติม 3', true, '550e8400-e29b-41d4-a716-446655440001', NOW(), NOW());

-- References for additional announcements
INSERT INTO course_materials (material_id, course_id, type, week, reference_id, reference_type, created_at, updated_at) VALUES
('f50e8400-e29b-41d4-a716-446655440004', '650e8400-e29b-41d4-a716-446655440001', 'announcement', 1, 'f50e8400-e29b-41d4-a716-446655440004', 'announcement', NOW(), NOW()),
('f50e8400-e29b-41d4-a716-446655440005', '650e8400-e29b-41d4-a716-446655440001', 'announcement', 2, 'f50e8400-e29b-41d4-a716-446655440005', 'announcement', NOW(), NOW()),
('f50e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', 'announcement', 3, 'f50e8400-e29b-41d4-a716-446655440006', 'announcement', NOW(), NOW());

-- =============================================
-- 12. QUEUE_JOBS (งานในคิว)
-- Note: Queues are now course-specific (e.g., code_execution_course_{courseID})
-- Each course has its own set of queues: code_execution, review, and file_processing
-- =============================================

INSERT INTO queue_jobs (id, type, status, user_id, material_id, course_id, data, result, created_at, updated_at, started_at, completed_at) VALUES
('g50e8400-e29b-41d4-a716-446655440001', 'code_execution', 'completed', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', '{"code": "def findMax(arr): return max(arr) if arr else None", "file_name": "solution.py", "material_id": "950e8400-e29b-41d4-a716-446655440006", "course_id": "650e8400-e29b-41d4-a716-446655440001"}', '{"success": true, "output": "Code executed successfully", "score": 100, "test_results": [{"test_case_id": "a50e8400-e29b-41d4-a716-446655440001", "passed": true, "input": "{\"arr\": [1, 5, 3, 9, 2]}", "expected": "{\"result\": 9}", "actual": "{\"result\": 9}"}]}', NOW(), NOW(), NOW() - INTERVAL '2 minutes', NOW() - INTERVAL '1 minute'),
('g50e8400-e29b-41d4-a716-446655440002', 'code_execution', 'completed', '550e8400-e29b-41d4-a716-446655440011', '950e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', '{"code": "def findMax(arr): return max(arr) if arr else None", "file_name": "solution.py", "material_id": "950e8400-e29b-41d4-a716-446655440006", "course_id": "650e8400-e29b-41d4-a716-446655440001"}', '{"success": true, "output": "Code executed successfully", "score": 100, "test_results": [{"test_case_id": "a50e8400-e29b-41d4-a716-446655440001", "passed": true, "input": "{\"arr\": [1, 5, 3, 9, 2]}", "expected": "{\"result\": 9}", "actual": "{\"result\": 9}"}]}', NOW(), NOW(), NOW() - INTERVAL '2 minutes', NOW() - INTERVAL '1 minute'),
('g50e8400-e29b-41d4-a716-446655440003', 'code_execution', 'completed', '550e8400-e29b-41d4-a716-446655440012', '950e8400-e29b-41d4-a716-446655440006', '650e8400-e29b-41d4-a716-446655440001', '{"code": "def findMax(arr): return arr[0] if arr else None", "file_name": "solution.py", "material_id": "950e8400-e29b-41d4-a716-446655440006", "course_id": "650e8400-e29b-41d4-a716-446655440001"}', '{"success": true, "output": "Code executed successfully", "score": 25, "test_results": [{"test_case_id": "a50e8400-e29b-41d4-a716-446655440001", "passed": false, "input": "{\"arr\": [1, 5, 3, 9, 2]}", "expected": "{\"result\": 9}", "actual": "{\"result\": 1}"}]}', NOW(), NOW(), NOW() - INTERVAL '2 minutes', NOW() - INTERVAL '1 minute'),
('g50e8400-e29b-41d4-a716-446655440004', 'code_review', 'pending', '550e8400-e29b-41d4-a716-446655440001', '950e8400-e29b-41d4-a716-446655440009', '650e8400-e29b-41d4-a716-446655440001', '{"material_id": "950e8400-e29b-41d4-a716-446655440009", "course_id": "650e8400-e29b-41d4-a716-446655440001", "review_notes": "Please review this PDF submission"}', NULL, NOW(), NOW(), NULL, NULL),
('g50e8400-e29b-41d4-a716-446655440005', 'code_review', 'pending', '550e8400-e29b-41d4-a716-446655440001', '950e8400-e29b-41d4-a716-446655440009', '650e8400-e29b-41d4-a716-446655440001', '{"material_id": "950e8400-e29b-41d4-a716-446655440009", "course_id": "650e8400-e29b-41d4-a716-446655440001", "review_notes": "Please review this PDF submission"}', NULL, NOW(), NOW(), NULL, NULL),
('g50e8400-e29b-41d4-a716-446655440006', 'code_execution', 'processing', '550e8400-e29b-41d4-a716-446655440013', '950e8400-e29b-41d4-a716-446655440011', '650e8400-e29b-41d4-a716-446655440001', '{"code": "def bubbleSort(arr): ...", "file_name": "solution.py", "material_id": "950e8400-e29b-41d4-a716-446655440011", "course_id": "650e8400-e29b-41d4-a716-446655440001"}', NULL, NOW(), NOW(), NOW() - INTERVAL '30 seconds', NULL),
('g50e8400-e29b-41d4-a716-446655440007', 'code_execution', 'failed', '550e8400-e29b-41d4-a716-446655440014', '950e8400-e29b-41d4-a716-446655440011', '650e8400-e29b-41d4-a716-446655440001', '{"code": "def bubbleSort(arr): ...", "file_name": "solution.py", "material_id": "950e8400-e29b-41d4-a716-446655440011", "course_id": "650e8400-e29b-41d4-a716-446655440001"}', NULL, NOW(), NOW(), NOW() - INTERVAL '1 minute', NOW() - INTERVAL '30 seconds');

-- =============================================
-- 13. STUDENT_COURSE_SCORES (คะแนนรวมของนักเรียนในคอร์ส)
-- =============================================

INSERT INTO student_course_scores (user_id, course_id, total_score, last_updated, created_at) VALUES
('550e8400-e29b-41d4-a716-446655440010', '650e8400-e29b-41d4-a716-446655440001', 250, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440011', '650e8400-e29b-41d4-a716-446655440001', 100, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440012', '650e8400-e29b-41d4-a716-446655440001', 25, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440013', '650e8400-e29b-41d4-a716-446655440001', 0, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440014', '650e8400-e29b-41d4-a716-446655440001', 0, NOW(), NOW());

-- =============================================
-- 14. EXERCISE_DRAFTS (ร่างงาน)
-- =============================================

INSERT INTO exercise_drafts (draft_id, user_id, material_id, code, file_name, file_path, file_size, created_at, updated_at) VALUES
('h50e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440010', '950e8400-e29b-41d4-a716-446655440006', 
'def findMax(arr):
    # Draft version
    max_val = 0
    for num in arr:
        if num > max_val:
            max_val = num
    return max_val', 'findMax_draft.py', '/drafts/findMax_draft.py', 256, NOW(), NOW()),
('h50e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440011', '950e8400-e29b-41d4-a716-446655440007', 
'class Stack:
    def __init__(self):
        self.items = []
    
    def push(self, item):
        self.items.append(item)
    
    def pop(self):
        return self.items.pop()', 'stack_draft.py', '/drafts/stack_draft.py', 512, NOW(), NOW());

-- =============================================
-- 15. COURSE_INVITATIONS (ลิงก์เชิญเข้าร่วมคอร์ส)
-- =============================================

INSERT INTO course_invitations (invitation_id, course_id, token, expires_at, created_at, updated_at) VALUES
('i50e8400-e29b-41d4-a716-446655440001', '650e8400-e29b-41d4-a716-446655440001', 'abc123def456ghi789jkl012mno345pqr678stu901vwx234', NOW() + INTERVAL '1 day', NOW(), NOW()),
('i50e8400-e29b-41d4-a716-446655440002', '650e8400-e29b-41d4-a716-446655440002', 'xyz789abc123def456ghi789jkl012mno345pqr678stu901', NOW() + INTERVAL '1 day', NOW(), NOW());

-- =============================================
-- สรุปข้อมูลตัวอย่าง
-- =============================================

/*
ข้อมูลตัวอย่างที่สร้างขึ้น:

1. USERS (10 คน)
   - ครู 3 คน
   - นักเรียน 5 คน
   - ผู้ช่วยสอน (TA) 2 คน

2. COURSES (3 คอร์ส)
   - คอร์สโครงสร้างข้อมูล (active)
   - คอร์สการเขียนโปรแกรมขั้นสูง (active)
   - คอร์สฐานข้อมูล (archived)

3. ENROLLMENTS (11 การลงทะเบียน)
   - ครู 3 คน
   - นักเรียน 6 คน
   - ผู้ช่วยสอน (TA) 2 คน

4. COURSE_WEEKS (5 สัปดาห์)
   - สำหรับคอร์สโครงสร้างข้อมูล

5. MATERIALS (13 เนื้อหา - แยกตามตาราง)
   - Documents: 3 ไฟล์
   - Videos: 2 ไฟล์
   - Code Exercises: 4 ข้อ (3 graded + 1 practice)
   - PDF Exercises: 2 ข้อ
   - Announcements: 3 ข้อ

5.1. COURSE_MATERIALS (13 references - ตารางกลาง)
   - References ไปยังตาราง materials ต่างๆ

6. TEST_CASES (10 กรณีทดสอบ)
   - สำหรับแบบฝึกหัดโค้ด 3 ข้อ

7. SUBMISSIONS (7 การส่งงาน)
   - การส่งงานโค้ด 4 ครั้ง
   - การส่งงาน PDF 3 ครั้ง

8. SUBMISSION_RESULTS (12 ผลการทดสอบ)
   - ผลการทดสอบโค้ด

9. STUDENT_PROGRESS (7 รายการ)
   - ความคืบหน้าของนักเรียน

10. VERIFICATION_LOGS (2 บันทึก)
    - การตรวจสอบงาน PDF

11. ANNOUNCEMENTS (3 ประกาศ)
    - ประกาศในคอร์ส

12. QUEUE_JOBS (7 งาน)
    - งานประมวลผลโค้ด (completed, processing, failed)
    - งานตรวจสอบ PDF (pending)
    - รองรับ course_id filtering
    - Queues are now course-specific: code_execution_course_{courseID}, review_course_{courseID}, file_processing_course_{courseID}

13. STUDENT_COURSE_SCORES (5 คะแนน)
    - คะแนนรวมของนักเรียน

14. EXERCISE_DRAFTS (2 ร่างงาน)
    - ร่างงานที่ยังไม่ส่ง

15. COURSE_INVITATIONS (2 ลิงก์เชิญ)
    - ลิงก์เชิญสำหรับคอร์สโครงสร้างข้อมูล (expires ใน 1 วัน)
    - ลิงก์เชิญสำหรับคอร์สการเขียนโปรแกรมขั้นสูง (expires ใน 1 วัน)

ข้อมูลนี้ครอบคลุมการใช้งานระบบทั้งหมดและแสดงให้เห็นถึงความสัมพันธ์ระหว่างตารางต่างๆ
รวมถึงการรองรับการกรอง queue jobs ตาม course_id สำหรับครูและผู้ช่วยสอน
*/
