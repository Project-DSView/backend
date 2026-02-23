package enums

// EnrollmentRole represents the role of a user in a course
type EnrollmentRole string

const (
	EnrollmentRoleStudent EnrollmentRole = "student"
	EnrollmentRoleTA      EnrollmentRole = "ta"
	EnrollmentRoleTeacher EnrollmentRole = "teacher" // Reserved for future use
)
