package types

// Repository Types
type CourseScoreStats struct {
	TotalStudents  int     `json:"total_students"`
	AverageScore   float64 `json:"average_score"`
	HighestScore   int     `json:"highest_score"`
	LowestScore    int     `json:"lowest_score"`
	PassedStudents int     `json:"passed_students"`
	FailedStudents int     `json:"failed_students"`
	CompletionRate float64 `json:"completion_rate"`
}

type StudentScoreStats struct {
	StudentID      string  `json:"student_id"`
	StudentName    string  `json:"student_name"`
	TotalScore     int     `json:"total_score"`
	MaxScore       int     `json:"max_score"`
	Percentage     float64 `json:"percentage"`
	CompletedCount int     `json:"completed_count"`
	TotalCount     int     `json:"total_count"`
	LastSubmitted  string  `json:"last_submitted,omitempty"`
}
