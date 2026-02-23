package auth

import (
	"fmt"

	"github.com/Project-DSView/backend/go/internal/domain/entities"
)

// Permission represents different levels of access
type Permission int

const (
	PermissionRead Permission = iota
	PermissionWrite
	PermissionDelete
	PermissionAdmin
)

// HasPermission checks if a user has a specific permission based on IsTeacher flag
func HasPermission(user *models.User, permission Permission) bool {
	if user == nil {
		return false
	}

	// Students (IsTeacher == false) → ไม่มีสิทธิ์
	if !user.IsTeacher {
		return false
	}

	// Teachers (IsTeacher == true) → ให้สิทธิ์อ่านอย่างเดียว
	if user.IsTeacher {
		return permission == PermissionRead
	}

	return false
}

// CanViewUsers checks if a user can view all users
func CanViewUsers(user *models.User) bool {
	return HasPermission(user, PermissionRead)
}

// CanUpdateUser checks if a user can update other users
func CanUpdateUser(user *models.User) bool {
	return HasPermission(user, PermissionWrite)
}

// CanDeleteUser checks if a user can delete other users
func CanDeleteUser(user *models.User) bool {
	return HasPermission(user, PermissionDelete)
}

// CanChangeUserRole checks if a user can change roles
func CanChangeUserRole(user *models.User) bool {
	return HasPermission(user, PermissionAdmin)
}

// ValidateRoleChange validates if a role change is allowed
func ValidateRoleChange(performer *models.User, target *models.User, newIsTeacher bool) error {
	if performer == nil || target == nil {
		return fmt.Errorf("invalid users")
	}

	// ต้องเป็น teacher เท่านั้นถึงจะเปลี่ยนได้
	if !performer.IsTeacher {
		return fmt.Errorf("insufficient permissions to change user roles")
	}

	// กันไม่ให้ teacher ธรรมดาไปแก้ตัวเองจนกลายเป็น student
	if performer.UserID == target.UserID && !newIsTeacher {
		return fmt.Errorf("teachers cannot demote themselves")
	}

	return nil
}

// FilterUsersByAccess filters users based on what the requesting user can see
func FilterUsersByAccess(users []models.User, requester *models.User) []models.User {
	if !CanViewUsers(requester) {
		return []models.User{} // No access
	}
	return users
}

// PermissionError represents a permission-related error
type PermissionError struct {
	Action   string
	Required Permission
	Current  bool // ใช้ isTeacher แทน role
}

// NewPermissionError creates a new permission error
func NewPermissionError(action string, required Permission, current bool) *PermissionError {
	return &PermissionError{
		Action:   action,
		Required: required,
		Current:  current,
	}
}
