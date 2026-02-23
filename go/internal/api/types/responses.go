package types

// Response Types
type CourseResponse struct {
	CourseID        string                 `json:"course_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	ImageURL        string                 `json:"image_url,omitempty"`
	CreatedBy       string                 `json:"created_by"`
	EnrollKey       string                 `json:"enroll_key,omitempty"`
	Status          string                 `json:"status"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
	EnrollmentCount int                    `json:"enrollment_count,omitempty"`
	ExerciseCount   int                    `json:"exercise_count,omitempty"`
	Creator         map[string]interface{} `json:"creator,omitempty"`
}

type EnrollmentResponse struct {
	EnrollmentID string `json:"enrollment_id"`
	CourseID     string `json:"course_id"`
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	FirstName    string `json:"firstname"`
	LastName     string `json:"lastname"`
	Email        string `json:"email"`
	EnrolledAt   string `json:"enrolled_at"`
}

type UserResponse struct {
	UserID     string `json:"user_id"`
	FirstName  string `json:"firstname"`
	LastName   string `json:"lastname"`
	Email      string `json:"email"`
	IsTeacher  bool   `json:"is_teacher"`
	ProfileImg string `json:"profile_img"`
}

type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type LoginResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
	Token   string       `json:"token,omitempty"`
}

type AuthURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
