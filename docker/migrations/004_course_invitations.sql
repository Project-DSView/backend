-- Migration: Create course_invitations table
-- Created: 2024-12-XX
-- Description: Creates table for course invitation links that allow students to enroll without enrollment key

BEGIN;

-- Create course_invitations table
CREATE TABLE IF NOT EXISTS course_invitations (
    invitation_id VARCHAR(36) PRIMARY KEY,
    course_id VARCHAR(36) NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_course_invitations_course FOREIGN KEY (course_id) REFERENCES courses(course_id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_course_invitations_course_id ON course_invitations(course_id);
CREATE INDEX IF NOT EXISTS idx_course_invitations_token ON course_invitations(token);
CREATE INDEX IF NOT EXISTS idx_course_invitations_expires_at ON course_invitations(expires_at);

COMMIT;

