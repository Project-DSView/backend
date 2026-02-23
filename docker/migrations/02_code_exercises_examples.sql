-- Code Exercises Examples for Data Structure Learning Platform
-- This file contains example SQL data for code-based exercises

-- =============================================
-- 1. HELLO WORLD EXERCISE
-- =============================================

-- Create Hello World exercise in code_exercises table
INSERT INTO code_exercises (
    material_id, 
    course_id, 
    title, 
    description, 
    week, 
    total_points, 
    deadline, 
    is_graded, 
    problem_statement, 
    example_inputs, 
    example_outputs, 
    constraints, 
    hints, 
    is_public, 
    created_by, 
    created_at, 
    updated_at
) VALUES (
    '950e8400-e29b-41d4-a716-446655440015', 
    '650e8400-e29b-41d4-a716-446655440001', 
    'แบบฝึกหัด: Hello World', 
    'แบบฝึกหัดพื้นฐานสำหรับการเขียนโปรแกรม Python', 
    1, 
    50, 
    '2025-12-31T23:59:59Z', 
    false, 
    'เขียนโปรแกรม Python ที่พิมพ์ข้อความ "Hello World" ออกมา

**ตัวอย่าง:**
- Input: ไม่มี
- Output: Hello World

**ข้อกำหนด:**
- ใช้ฟังก์ชัน print() เท่านั้น
- ไม่ต้องรับ input จากผู้ใช้', 
    '[]'::jsonb, 
    '["Hello World"]'::jsonb, 
    'ไม่มีข้อจำกัด', 
    'ใช้ฟังก์ชัน print() พิมพ์ข้อความออกมา', 
    true, 
    '550e8400-e29b-41d4-a716-446655440001', 
    NOW(), 
    NOW()
);

-- Create CourseMaterial reference for Hello World exercise
INSERT INTO course_materials (
    material_id, 
    course_id, 
    type, 
    week, 
    reference_id, 
    reference_type, 
    created_at, 
    updated_at
) VALUES (
    '950e8400-e29b-41d4-a716-446655440015', 
    '650e8400-e29b-41d4-a716-446655440001', 
    'code_exercise', 
    1, 
    '950e8400-e29b-41d4-a716-446655440015', 
    'code_exercise', 
    NOW(), 
    NOW()
);

-- Test cases for Hello World exercise
INSERT INTO test_cases (
    test_case_id, 
    material_id, 
    input_data, 
    expected_output, 
    is_public, 
    display_name, 
    created_at, 
    updated_at
) VALUES (
    'a50e8400-e29b-41d4-a716-446655440011', 
    '950e8400-e29b-41d4-a716-446655440015', 
    '{}', 
    '{"output": "Hello World"}', 
    true, 
    'Test Case 1: Basic Hello World', 
    NOW(), 
    NOW()
);

-- =============================================
-- 2. SINGLY LINKED LIST EXERCISE
-- =============================================

-- Create Singly Linked List exercise in code_exercises table
INSERT INTO code_exercises (
    material_id, 
    course_id, 
    title, 
    description, 
    week, 
    total_points, 
    deadline, 
    is_graded, 
    problem_statement, 
    example_inputs, 
    example_outputs, 
    constraints, 
    hints, 
    is_public, 
    created_by, 
    created_at, 
    updated_at
) VALUES (
    '950e8400-e29b-41d4-a716-446655440016', 
    '650e8400-e29b-41d4-a716-446655440001', 
    'แบบฝึกหัด: Singly Linked List Implementation', 
    'สร้าง Singly Linked List class พร้อมฟังก์ชันพื้นฐาน', 
    2, 
    200, 
    '2025-12-31T23:59:59Z', 
    true, 
    'สร้าง Singly Linked List class ที่มีฟังก์ชันต่อไปนี้:

1. `insertFront(data)` - เพิ่มข้อมูลที่ด้านหน้า
2. `insertLast(data)` - เพิ่มข้อมูลที่ด้านหลัง  
3. `insertBefore(target, data)` - เพิ่มข้อมูลก่อน target
4. `delete(data)` - ลบข้อมูลที่ระบุ
5. `traverse()` - แสดงข้อมูลทั้งหมด

**ตัวอย่าง:**
```python
mylist = SinglyLinkedList()
mylist.insertFront("Tony")
mylist.insertFront("John")
mylist.traverse()  # -> John -> Tony
mylist.insertLast("Saori")
mylist.traverse()  # -> John -> Tony -> Saori
mylist.insertBefore("Tony", "Ako")
mylist.traverse()  # -> John -> Ako -> Tony -> Saori
mylist.delete("John")
mylist.traverse()  # -> Ako -> Tony -> Saori
```

**ข้อกำหนด:**
- ใช้ class และ method ตามที่กำหนด
- ใช้ print() ในการแสดงผล traverse
- จัดการกรณีที่ list ว่าง', 
    '["insertFront(\"Tony\"), insertFront(\"John\"), traverse()", "insertLast(\"Saori\"), traverse()", "insertBefore(\"Tony\", \"Ako\"), traverse()", "delete(\"John\"), traverse()"]'::jsonb, 
    '["-> John -> Tony", "-> John -> Tony -> Saori", "-> John -> Ako -> Tony -> Saori", "-> Ako -> Tony -> Saori"]'::jsonb, 
    '1 ≤ operations ≤ 100
1 ≤ data length ≤ 50', 
    'สร้าง Node class สำหรับเก็บข้อมูลและ pointer ไปยัง node ถัดไป', 
    true, 
    '550e8400-e29b-41d4-a716-446655440001', 
    NOW(), 
    NOW()
);

-- Create CourseMaterial reference for Singly Linked List exercise
INSERT INTO course_materials (
    material_id, 
    course_id, 
    type, 
    week, 
    reference_id, 
    reference_type, 
    created_at, 
    updated_at
) VALUES (
    '950e8400-e29b-41d4-a716-446655440016', 
    '650e8400-e29b-41d4-a716-446655440001', 
    'code_exercise', 
    2, 
    '950e8400-e29b-41d4-a716-446655440016', 
    'code_exercise', 
    NOW(), 
    NOW()
);

-- Test cases for Singly Linked List exercise
INSERT INTO test_cases (
    test_case_id, 
    material_id, 
    input_data, 
    expected_output, 
    is_public, 
    display_name, 
    created_at, 
    updated_at
) VALUES (
    'a50e8400-e29b-41d4-a716-446655440012', 
    '950e8400-e29b-41d4-a716-446655440016', 
    '{"operations": ["insertFront(\"Tony\")", "insertFront(\"John\")", "traverse()"]}', 
    '{"output": "-> John -> Tony"}', 
    false, 
    'Test Case 1: Basic insertFront and traverse', 
    NOW(), 
    NOW()
),
(
    'a50e8400-e29b-41d4-a716-446655440013', 
    '950e8400-e29b-41d4-a716-446655440016', 
    '{"operations": ["insertFront(\"Tony\")", "insertLast(\"Saori\")", "traverse()"]}', 
    '{"output": "-> Tony -> Saori"}', 
    false, 
    'Test Case 2: insertFront and insertLast', 
    NOW(), 
    NOW()
),
(
    'a50e8400-e29b-41d4-a716-446655440014', 
    '950e8400-e29b-41d4-a716-446655440016', 
    '{"operations": ["insertFront(\"Tony\")", "insertFront(\"John\")", "insertBefore(\"Tony\", \"Ako\")", "traverse()"]}', 
    '{"output": "-> John -> Ako -> Tony"}', 
    false, 
    'Test Case 3: insertBefore operation', 
    NOW(), 
    NOW()
),
(
    'a50e8400-e29b-41d4-a716-446655440015', 
    '950e8400-e29b-41d4-a716-446655440016', 
    '{"operations": ["insertFront(\"Tony\")", "insertFront(\"John\")", "delete(\"John\")", "traverse()"]}', 
    '{"output": "-> Tony"}', 
    false, 
    'Test Case 4: delete operation', 
    NOW(), 
    NOW()
),
(
    'a50e8400-e29b-41d4-a716-446655440016', 
    '950e8400-e29b-41d4-a716-446655440016', 
    '{"operations": ["traverse()"]}', 
    '{"output": "This is an empty list."}', 
    true, 
    'Test Case 5: Empty list traverse', 
    NOW(), 
    NOW()
),
(
    'a50e8400-e29b-41d4-a716-446655440017', 
    '950e8400-e29b-41d4-a716-446655440016', 
    '{"operations": ["insertFront(\"Tony\")", "insertFront(\"John\")", "insertLast(\"Saori\")", "insertBefore(\"Tony\", \"Ako\")", "delete(\"John\")", "traverse()"]}', 
    '{"output": "-> Ako -> Tony -> Saori"}', 
    true, 
    'Test Case 6: Complete workflow', 
    NOW(), 
    NOW()
);

-- =============================================
-- 3. EXAMPLE SUBMISSIONS
-- =============================================

-- Example submission for Hello World
INSERT INTO submissions (
    submission_id, 
    user_id, 
    material_id, 
    code, 
    passed_count, 
    failed_count, 
    total_score, 
    status, 
    is_late_submission, 
    submitted_at
) VALUES (
    'b50e8400-e29b-41d4-a716-446655440009', 
    '550e8400-e29b-41d4-a716-446655440010', 
    '950e8400-e29b-41d4-a716-446655440015', 
    'print("Hello World")', 
    1, 
    0, 
    50, 
    'completed', 
    false, 
    NOW()
);

-- Example submission for Singly Linked List
INSERT INTO submissions (
    submission_id, 
    user_id, 
    material_id, 
    code, 
    passed_count, 
    failed_count, 
    total_score, 
    status, 
    is_late_submission, 
    submitted_at
) VALUES (
    'b50e8400-e29b-41d4-a716-446655440010', 
    '550e8400-e29b-41d4-a716-446655440010', 
    '950e8400-e29b-41d4-a716-446655440016', 
    'class DataNode:
    def __init__(self, name):
        self.name = name
        self.next = None

class SinglyLinkedList:
    def __init__(self):
        self.count = 0
        self.head = None

    def traverse(self):
        start = self.head
        while start != None:
            print("->", start.name, end=" ")
            start = start.next
        if self.head == None:
            print("This is an empty list.")
        print()
    
    def insertFront(self, name):
        pNew = DataNode(name)
        if self.head == None:
            self.head = pNew
        else:
            pNew.next = self.head
            self.head = pNew

    def insertLast(self, name):
        pNew = DataNode(name)
        if self.head == None:
            self.head = pNew
        else:
            start = self.head
            while start.next != None:
                start = start.next
            start.next = pNew
    
    def insertBefore(self, Node, name):
        pNew = DataNode(name)
        start = self.head
        if self.head.name == Node:
            pNew.next = self.head
            self.head = pNew
        else:
            while start.next != None:
                if start.next.name == Node:
                    pNew.next = start.next
                    start.next = pNew
                    return
                start = start.next
            print("Cannot insert, <" + Node + "> does not exist.")

    def delete(self, name):
        if self.head.name == name:
            self.head = self.head.next
        else:
            start = self.head
            while start.next != None:
                if start.next.name == name:
                    start.next = start.next.next
                    return
                start = start.next
            print("Cannot delete, <" + name + "> does not exist.")', 
    6, 
    0, 
    200, 
    'completed', 
    false, 
    NOW()
);

-- =============================================
-- 4. SUBMISSION RESULTS
-- =============================================

-- Results for Hello World submission
INSERT INTO submission_results (
    result_id, 
    submission_id, 
    test_case_id, 
    status, 
    actual_output, 
    created_at
) VALUES (
    'c50e8400-e29b-41d4-a716-446655440013', 
    'b50e8400-e29b-41d4-a716-446655440009', 
    'a50e8400-e29b-41d4-a716-446655440011', 
    'passed', 
    '{"output": "Hello World"}', 
    NOW()
);

-- Results for Singly Linked List submission
INSERT INTO submission_results (
    result_id, 
    submission_id, 
    test_case_id, 
    status, 
    actual_output, 
    created_at
) VALUES (
    'c50e8400-e29b-41d4-a716-446655440014', 
    'b50e8400-e29b-41d4-a716-446655440010', 
    'a50e8400-e29b-41d4-a716-446655440012', 
    'passed', 
    '{"output": "-> John -> Tony"}', 
    NOW()
),
(
    'c50e8400-e29b-41d4-a716-446655440015', 
    'b50e8400-e29b-41d4-a716-446655440010', 
    'a50e8400-e29b-41d4-a716-446655440013', 
    'passed', 
    '{"output": "-> Tony -> Saori"}', 
    NOW()
),
(
    'c50e8400-e29b-41d4-a716-446655440016', 
    'b50e8400-e29b-41d4-a716-446655440010', 
    'a50e8400-e29b-41d4-a716-446655440014', 
    'passed', 
    '{"output": "-> John -> Ako -> Tony"}', 
    NOW()
),
(
    'c50e8400-e29b-41d4-a716-446655440017', 
    'b50e8400-e29b-41d4-a716-446655440010', 
    'a50e8400-e29b-41d4-a716-446655440015', 
    'passed', 
    '{"output": "-> Tony"}', 
    NOW()
),
(
    'c50e8400-e29b-41d4-a716-446655440018', 
    'b50e8400-e29b-41d4-a716-446655440010', 
    'a50e8400-e29b-41d4-a716-446655440016', 
    'passed', 
    '{"output": "This is an empty list."}', 
    NOW()
),
(
    'c50e8400-e29b-41d4-a716-446655440019', 
    'b50e8400-e29b-41d4-a716-446655440010', 
    'a50e8400-e29b-41d4-a716-446655440017', 
    'passed', 
    '{"output": "-> Ako -> Tony -> Saori"}', 
    NOW()
);

-- =============================================
-- 5. STUDENT PROGRESS
-- =============================================

INSERT INTO student_progress (
    progress_id, 
    user_id, 
    material_id, 
    status, 
    score, 
    seat_number, 
    last_submitted_at, 
    created_at, 
    updated_at
) VALUES (
    'd50e8400-e29b-41d4-a716-446655440008', 
    '550e8400-e29b-41d4-a716-446655440010', 
    '950e8400-e29b-41d4-a716-446655440015', 
    'completed', 
    50, 
    'A001', 
    NOW(), 
    NOW(), 
    NOW()
),
(
    'd50e8400-e29b-41d4-a716-446655440009', 
    '550e8400-e29b-41d4-a716-446655440010', 
    '950e8400-e29b-41d4-a716-446655440016', 
    'completed', 
    200, 
    'A001', 
    NOW(), 
    NOW(), 
    NOW()
);

-- =============================================
-- Summary
-- =============================================

/*
Created example data for:

1. HELLO WORLD EXERCISE
   - Material ID: 950e8400-e29b-41d4-a716-446655440015
   - 1 test case (public)
   - Simple print("Hello World") test

2. SINGLY LINKED LIST EXERCISE  
   - Material ID: 950e8400-e29b-41d4-a716-446655440016
   - 6 test cases (2 public, 4 private)
   - Complete class implementation test
   - Tests: insertFront, insertLast, insertBefore, delete, traverse

3. EXAMPLE SUBMISSIONS
   - Working submissions for both exercises
   - Complete submission results

4. STUDENT PROGRESS
   - Progress tracking for both exercises

These examples can be used to test the code submission feature.
*/
