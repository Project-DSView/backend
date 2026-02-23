-- Migration: Separate course_materials into videos, documents, code_exercises, pdf_exercises
-- Created: 2024-01-XX
-- Description: Migrates data from course_materials table to separate tables based on type
-- This file is for Docker-based migrations

BEGIN;

-- Step 1: Create new tables

-- Videos table
CREATE TABLE IF NOT EXISTS videos (
    material_id VARCHAR(36) PRIMARY KEY,
    course_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    week INT NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    video_url TEXT NOT NULL,
    CONSTRAINT fk_videos_course FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE,
    CONSTRAINT fk_videos_creator FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

-- Documents table
CREATE TABLE IF NOT EXISTS documents (
    material_id VARCHAR(36) PRIMARY KEY,
    course_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    week INT NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    file_url TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT DEFAULT 0,
    mime_type VARCHAR(100),
    CONSTRAINT fk_documents_course FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE,
    CONSTRAINT fk_documents_creator FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

-- Code exercises table
CREATE TABLE IF NOT EXISTS code_exercises (
    material_id VARCHAR(36) PRIMARY KEY,
    course_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    week INT NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    total_points INT NOT NULL,
    deadline VARCHAR(50),
    is_graded BOOLEAN DEFAULT true,
    problem_statement TEXT NOT NULL,
    problem_images JSONB DEFAULT '[]'::jsonb,
    example_inputs JSONB DEFAULT '[]'::jsonb,
    example_outputs JSONB DEFAULT '[]'::jsonb,
    constraints TEXT,
    hints TEXT,
    CONSTRAINT fk_code_exercises_course FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE,
    CONSTRAINT fk_code_exercises_creator FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

-- PDF exercises table
CREATE TABLE IF NOT EXISTS pdf_exercises (
    material_id VARCHAR(36) PRIMARY KEY,
    course_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    week INT NOT NULL DEFAULT 0,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    total_points INT NOT NULL,
    deadline VARCHAR(50),
    is_graded BOOLEAN DEFAULT true,
    file_url TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT DEFAULT 0,
    mime_type VARCHAR(100),
    CONSTRAINT fk_pdf_exercises_course FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE,
    CONSTRAINT fk_pdf_exercises_creator FOREIGN KEY (created_by) REFERENCES users(user_id) ON DELETE RESTRICT
);

-- Step 2: Create indexes
CREATE INDEX IF NOT EXISTS idx_videos_course_id ON videos(course_id);
CREATE INDEX IF NOT EXISTS idx_videos_week ON videos(week);
CREATE INDEX IF NOT EXISTS idx_videos_created_by ON videos(created_by);

CREATE INDEX IF NOT EXISTS idx_documents_course_id ON documents(course_id);
CREATE INDEX IF NOT EXISTS idx_documents_week ON documents(week);
CREATE INDEX IF NOT EXISTS idx_documents_created_by ON documents(created_by);

CREATE INDEX IF NOT EXISTS idx_code_exercises_course_id ON code_exercises(course_id);
CREATE INDEX IF NOT EXISTS idx_code_exercises_week ON code_exercises(week);
CREATE INDEX IF NOT EXISTS idx_code_exercises_created_by ON code_exercises(created_by);

CREATE INDEX IF NOT EXISTS idx_pdf_exercises_course_id ON pdf_exercises(course_id);
CREATE INDEX IF NOT EXISTS idx_pdf_exercises_week ON pdf_exercises(week);
CREATE INDEX IF NOT EXISTS idx_pdf_exercises_created_by ON pdf_exercises(created_by);

-- Step 3: Migrate data from course_materials to new tables

-- Migrate videos
INSERT INTO videos (
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at, video_url
)
SELECT 
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at, video_url
FROM course_materials
WHERE type = 'video' AND video_url IS NOT NULL AND video_url != ''
ON CONFLICT (material_id) DO NOTHING;

-- Migrate documents
INSERT INTO documents (
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at,
    file_url, file_name, file_size, mime_type
)
SELECT 
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at,
    file_url, file_name, file_size, mime_type
FROM course_materials
WHERE type = 'document' AND file_url IS NOT NULL AND file_url != ''
ON CONFLICT (material_id) DO NOTHING;

-- Migrate code exercises
INSERT INTO code_exercises (
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at,
    total_points, deadline, is_graded, problem_statement, problem_images, example_inputs, example_outputs, constraints, hints
)
SELECT 
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at,
    total_points, deadline, is_graded, problem_statement, problem_images, example_inputs, example_outputs, constraints, hints
FROM course_materials
WHERE type = 'code_exercise' AND total_points IS NOT NULL
ON CONFLICT (material_id) DO NOTHING;

-- Migrate PDF exercises
INSERT INTO pdf_exercises (
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at,
    total_points, deadline, is_graded, file_url, file_name, file_size, mime_type
)
SELECT 
    material_id, course_id, title, description, week, is_public, created_by, created_at, updated_at,
    total_points, deadline, is_graded, file_url, file_name, file_size, mime_type
FROM course_materials
WHERE type = 'pdf_exercise' AND total_points IS NOT NULL AND file_url IS NOT NULL AND file_url != ''
ON CONFLICT (material_id) DO NOTHING;

-- Step 4: Update foreign keys in related tables to use polymorphic association
-- Add material_type column to track which table the material is in

-- Add material_type column to submissions if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'submissions' AND column_name = 'material_type') THEN
        ALTER TABLE submissions ADD COLUMN material_type VARCHAR(20);
    END IF;
END $$;

-- Update material_type in submissions based on existing data
UPDATE submissions s
SET material_type = (
    SELECT type FROM course_materials cm WHERE cm.material_id = s.material_id
)
WHERE material_type IS NULL;

-- Add material_type column to test_cases if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'test_cases' AND column_name = 'material_type') THEN
        ALTER TABLE test_cases ADD COLUMN material_type VARCHAR(20);
    END IF;
END $$;

-- Update material_type in test_cases (only code_exercises have test cases)
UPDATE test_cases tc
SET material_type = 'code_exercise'
WHERE material_type IS NULL AND material_id IS NOT NULL;

-- Add material_type column to queue_jobs if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'queue_jobs' AND column_name = 'material_type') THEN
        ALTER TABLE queue_jobs ADD COLUMN material_type VARCHAR(20);
    END IF;
END $$;

-- Update material_type in queue_jobs
UPDATE queue_jobs qj
SET material_type = (
    SELECT type FROM course_materials cm WHERE cm.material_id = qj.material_id
)
WHERE material_type IS NULL AND material_id IS NOT NULL;

-- Add material_type column to student_progress if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'student_progress' AND column_name = 'material_type') THEN
        ALTER TABLE student_progress ADD COLUMN material_type VARCHAR(20);
    END IF;
END $$;

-- Update material_type in student_progress
UPDATE student_progress sp
SET material_type = (
    SELECT type FROM course_materials cm WHERE cm.material_id = sp.material_id
)
WHERE material_type IS NULL AND material_id IS NOT NULL;

-- Add material_type column to exercise_drafts if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'exercise_drafts' AND column_name = 'material_type') THEN
        ALTER TABLE exercise_drafts ADD COLUMN material_type VARCHAR(20);
    END IF;
END $$;

-- Update material_type in exercise_drafts
UPDATE exercise_drafts ed
SET material_type = (
    SELECT type FROM course_materials cm WHERE cm.material_id = ed.material_id
)
WHERE material_type IS NULL AND material_id IS NOT NULL;

-- Step 5: Create indexes for material_type columns
CREATE INDEX IF NOT EXISTS idx_submissions_material_type ON submissions(material_type);
CREATE INDEX IF NOT EXISTS idx_test_cases_material_type ON test_cases(material_type);
CREATE INDEX IF NOT EXISTS idx_queue_jobs_material_type ON queue_jobs(material_type);
CREATE INDEX IF NOT EXISTS idx_student_progress_material_type ON student_progress(material_type);
CREATE INDEX IF NOT EXISTS idx_exercise_drafts_material_type ON exercise_drafts(material_type);

COMMIT;

-- Note: The old course_materials table will be kept for backward compatibility
-- It can be dropped in a future migration after verifying all systems work correctly


















