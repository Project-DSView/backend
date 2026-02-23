package auth

// CanViewCourse checks if a user can view courses
func CanViewCourse() bool {
	return true // All authenticated users can view courses
}

// CanCreateCourse checks if a user can create courses
func CanCreateCourse(isTeacher bool) bool {
	return isTeacher
}

// CanEditCourse checks if a user can edit a course
func CanEditCourse(isTeacher bool, createdBy, userID string) bool {
	if isTeacher {
		return true
	}
	// Students cannot edit courses
	return false
}

// CanDeleteCourse checks if a user can delete a course
func CanDeleteCourse(isTeacher bool, createdBy, userID string) bool {
	return CanEditCourse(isTeacher, createdBy, userID)
}

// CanViewEnrollments checks if a user can view course enrollments
func CanViewEnrollments(isTeacher bool) bool {
	return true
}

// CanEnrollInCourse checks if a user can enroll in courses
func CanEnrollInCourse() bool {
	return true // All users can enroll
}

// CanManageEnrollmentRoles checks if user can change enrollment roles
func CanManageEnrollmentRoles(isTeacher bool) bool {
	return isTeacher
}

// CanViewCourseScores checks if user can view scores in a course
func CanViewCourseScores(isTeacher bool, enrollmentRole string) bool {
	// Teachers can view all course scores
	if isTeacher {
		return true
	}
	// TAs can view course scores
	if enrollmentRole == "ta" || enrollmentRole == "teacher" {
		return true
	}
	return false
}

// CanViewStudentScores checks if user can view specific student scores
func CanViewStudentScores(isTeacher bool, enrollmentRole string, targetUserID, currentUserID string) bool {
	// Users can always view their own scores
	if targetUserID == currentUserID {
		return true
	}
	// Teachers and TAs can view any student's scores
	return CanViewCourseScores(isTeacher, enrollmentRole)
}
