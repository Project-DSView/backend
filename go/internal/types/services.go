package types

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service Types
type SubmitResult struct {
	Submission interface{} `json:"submission"`
	Results    interface{} `json:"results"`
}

type SubmissionFilter struct {
	UserID     string `json:"user_id,omitempty"`
	MaterialID string `json:"material_id,omitempty"`
	Status     string `json:"status,omitempty"`
	CourseID   string `json:"course_id,omitempty"`
	Page       int    `json:"page,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type QueueJobData struct {
	Code           string      `json:"code,omitempty"`
	FileName       string      `json:"file_name,omitempty"`
	FileURL        string      `json:"file_url,omitempty"`
	FileSize       int64       `json:"file_size,omitempty"`
	SubmissionType string      `json:"submission_type,omitempty"` // "code" or "pdf"
	MaterialID     string      `json:"material_id,omitempty"`
	CourseID       string      `json:"course_id,omitempty"`
	SubmissionID   string      `json:"submission_id,omitempty"`
	LabRoom        string      `json:"lab_room,omitempty"`
	TableNumber    string      `json:"table_number,omitempty"`
	TestCases      interface{} `json:"test_cases,omitempty"`
	Language       string      `json:"language,omitempty"`
	ReviewNotes    string      `json:"review_notes,omitempty"`
}

type QueueJobResult struct {
	Success     bool                   `json:"success"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	TestResults []TestResult           `json:"test_results,omitempty"`
	Score       float64                `json:"score,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type TestResult struct {
	TestCaseID string `json:"test_case_id"`
	Passed     bool   `json:"passed"`
	Input      string `json:"input"`
	Expected   string `json:"expected"`
	Actual     string `json:"actual"`
	Error      string `json:"error,omitempty"`
}

type UserStatistics struct {
	TotalUsers   int64     `json:"total_users"`
	TeacherCount int64     `json:"teacher_count"`
	StudentCount int64     `json:"student_count"`
	ActiveUsers  int64     `json:"active_users"`
	RecentUsers  int64     `json:"recent_users"`
	LastUpdated  time.Time `json:"last_updated"`
}

type CourseProgressRow struct {
	UserID      string `json:"user_id"`
	MaterialID  string `json:"material_id"`
	Status      string `json:"status"`
	Score       int    `json:"score"`
	SubmittedAt string `json:"submitted_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsTeacher bool   `json:"is_teacher"`
	jwt.RegisteredClaims
}
