package enums

// CourseStatus represents the status of a course
type CourseStatus string

const (
	CourseStatusActive   CourseStatus = "active"
	CourseStatusArchived CourseStatus = "archived"
)

// IsValidCourseStatus checks if the status is valid
func IsValidCourseStatus(status string) bool {
	switch CourseStatus(status) {
	case CourseStatusActive, CourseStatusArchived:
		return true
	default:
		return false
	}
}
