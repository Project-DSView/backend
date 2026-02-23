package response

import (
	"strings"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/domain/enums"
)

// CourseResponse represents the course data returned to the client
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
	MaterialCount   int                    `json:"material_count,omitempty"`
	Creator         map[string]interface{} `json:"creator,omitempty"`
}

// EnrollmentResponse represents enrollment data returned to the client
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

// UserResponse represents the user data returned to the client
type UserResponse struct {
	UserID     string `json:"user_id"`
	FirstName  string `json:"firstname"`
	LastName   string `json:"lastname"`
	Email      string `json:"email"`
	IsTeacher  bool   `json:"is_teacher"`
	ProfileImg string `json:"profile_img"`
}

// ConvertToCourseResponse converts a models.Course to CourseResponse
func ConvertToCourseResponse(course *models.Course, includeEnrollKey bool) CourseResponse {
	response := CourseResponse{
		CourseID:        course.CourseID,
		Name:            course.Name,
		Description:     course.Description,
		ImageURL:        course.ImageURL,
		CreatedBy:       course.CreatedBy,
		Status:          string(course.Status),
		CreatedAt:       course.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       course.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		EnrollmentCount: course.EnrollmentCount,
		MaterialCount:   course.MaterialCount,
	}

	if includeEnrollKey {
		response.EnrollKey = course.EnrollKey
	}

	if course.CreatorInfo != nil {
		response.Creator = course.CreatorInfo.ToJSON()
	}

	return response
}

// ConvertToEnrollmentResponse converts a models.Enrollment to EnrollmentResponse
func ConvertToEnrollmentResponse(enrollment *models.Enrollment) EnrollmentResponse {
	response := EnrollmentResponse{
		EnrollmentID: enrollment.EnrollmentID,
		CourseID:     enrollment.CourseID,
		UserID:       enrollment.UserID,
		Role:         string(enrollment.Role),
		EnrolledAt:   enrollment.EnrolledAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if enrollment.UserInfo != nil {
		response.FirstName = enrollment.UserInfo.FirstName
		response.LastName = enrollment.UserInfo.LastName
		response.Email = enrollment.UserInfo.Email
	}

	return response
}

// ConvertToUserResponse converts a models.User to UserResponse
func ConvertToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		UserID:     user.UserID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		IsTeacher:  user.IsTeacher,
		ProfileImg: user.ProfileImg,
	}
}

// BuildCourseUpdates builds updates map for database operations
func BuildCourseUpdates(name, description, status, enrollKey *string) map[string]interface{} {
	updates := make(map[string]interface{})

	if name != nil {
		updates["name"] = SanitizeInput(*name)
	}
	if description != nil {
		updates["description"] = SanitizeInput(*description)
	}
	if status != nil && enums.IsValidCourseStatus(*status) {
		updates["status"] = *status
	}
	if enrollKey != nil {
		updates["enroll_key"] = SanitizeInput(*enrollKey)
	}

	return updates
}

// SanitizeInput sanitizes user input
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}
