package types

// Course Requests
type CreateCourseRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required,min=1"`
	ImageURL    string `json:"image_url,omitempty"`
}

type UpdateCourseRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=1"`
	ImageURL    *string `json:"image_url,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active archived"`
}

// Exercise Requests
type CreateExerciseRequest struct {
	Title           string                  `json:"title" validate:"required,min=1,max=255"`
	Description     string                  `json:"description" validate:"required,min=1"`
	DataTypeAllowed int                     `json:"data_type_allowed" validate:"required,min=1"`
	TotalPoints     string                  `json:"total_points" validate:"required"`
	Deadline        *string                 `json:"deadline,omitempty"`
	CourseIDs       []string                `json:"course_ids,omitempty"`
	TestCases       []CreateTestCaseRequest `json:"test_cases,omitempty"`
}

type CreateTestCaseRequest struct {
	InputData      map[string]interface{} `json:"input_data" validate:"required"`
	ExpectedOutput map[string]interface{} `json:"expected_output" validate:"required"`
}

type UpdateExerciseRequest struct {
	Title           *string  `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description     *string  `json:"description,omitempty" validate:"omitempty,min=1"`
	DataTypeAllowed *int     `json:"data_type_allowed,omitempty" validate:"omitempty,min=1"`
	TotalPoints     *string  `json:"total_points,omitempty"`
	Status          *string  `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
	Deadline        *string  `json:"deadline,omitempty"`
	CourseIDs       []string `json:"course_ids,omitempty"`
}

// Course Material Requests
type CreateCourseMaterialRequest struct {
	CourseID    string `json:"course_id" validate:"required"`
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"required,min=1"`
	Type        string `json:"type" validate:"required,oneof=document code_exercise pdf_exercise"`
	Week        int    `json:"week" validate:"min=0"`
	FileURL     string `json:"file_url,omitempty"`
	VideoURL    string `json:"video_url,omitempty"`
	FileName    string `json:"file_name,omitempty"`
	FileSize    *int64 `json:"file_size,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
	IsPublic    bool   `json:"is_public"`

	// Exercise-specific fields
	ExerciseID     *string `json:"exercise_id,omitempty"`
	TotalPoints    *int    `json:"total_points,omitempty"`
	Deadline       *string `json:"deadline,omitempty"`
	SubmissionType string  `json:"submission_type,omitempty" validate:"omitempty,oneof=file code"`
}

type UpdateCourseMaterialRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=1"`
	Type        *string `json:"type,omitempty" validate:"omitempty,oneof=document code_exercise pdf_exercise"`
	Week        *int    `json:"week,omitempty" validate:"omitempty,min=0"`
	FileURL     *string `json:"file_url,omitempty"`
	VideoURL    *string `json:"video_url,omitempty"`
	FileName    *string `json:"file_name,omitempty"`
	FileSize    *int64  `json:"file_size,omitempty"`
	MimeType    *string `json:"mime_type,omitempty"`
	IsPublic    *bool   `json:"is_public,omitempty"`

	// Exercise-specific fields
	ExerciseID     *string `json:"exercise_id,omitempty"`
	TotalPoints    *int    `json:"total_points,omitempty"`
	Deadline       *string `json:"deadline,omitempty"`
	SubmissionType *string `json:"submission_type,omitempty" validate:"omitempty,oneof=file code"`

	// Code exercise-specific fields
	ProblemStatement *string                `json:"problem_statement,omitempty"`
	Constraints      *string                `json:"constraints,omitempty"`
	Hints            *string                `json:"hints,omitempty"`
	TestCases        *[]map[string]interface{} `json:"test_cases,omitempty"`

	// Announcement-specific fields
	Content *string `json:"content,omitempty"`
}

// PDF Exercise Submission Requests
type ApprovePDFSubmissionRequest struct {
	Score   int    `json:"score" validate:"required,min=0,max=100"`
	Comment string `json:"comment" validate:"required,min=1"`
}

type RejectPDFSubmissionRequest struct {
	Comment string `json:"comment" validate:"required,min=1"`
}

// Announcement Requests
type CreateAnnouncementRequest struct {
	CourseID string `json:"course_id" validate:"required"`
	Title    string `json:"title" validate:"required,min=1,max=255"`
	Content  string `json:"content" validate:"required,min=1"`
}

type UpdateAnnouncementRequest struct {
	Title   *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Content *string `json:"content,omitempty" validate:"omitempty,min=1"`
}

// Playground Requests
type PlaygroundRunRequest struct {
	Code     string `json:"code" validate:"required"`
	DataType string `json:"dataType" validate:"required"`
}

// Course Score Requests
type CourseScoreBatchUpdateRequest struct {
	CourseID string   `json:"course_id" validate:"required"`
	UserIDs  []string `json:"user_ids" validate:"required,min=1"`
}

// Test Requests
type TestTokenRequest struct {
	IsTeacher bool   `json:"is_teacher"`
	Email     string `json:"email,omitempty"`
	Name      string `json:"name,omitempty"`
	Duration  string `json:"duration,omitempty"` // Optional custom duration like "1h", "30m", "24h"
}
