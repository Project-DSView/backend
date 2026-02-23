package services

import (
	"strings"
	"time"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/types"
	"github.com/Project-DSView/backend/go/pkg/errors"
	"github.com/Project-DSView/backend/go/pkg/validation"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

// UserStatistics represents user statistics data
// UserStatistics is now defined in internal/types/services.go

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) CreateUser(googleUser *types.GoogleUser) (*models.User, error) {
	// Parse first name and last name from Google's name field
	firstName, lastName := validation.ParseFullName(googleUser.Name, googleUser.GivenName, googleUser.FamilyName)

	userModel := &models.User{
		FirstName:  firstName,
		LastName:   lastName,
		Email:      googleUser.Email,
		IsTeacher:  false, // Default to student
		ProfileImg: googleUser.Picture,
	}

	if err := s.db.Create(userModel).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create user")
	}

	return userModel, nil
}

func (s *UserService) CreateOrUpdateUser(googleUser *types.GoogleUser) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.GetUserByEmail(googleUser.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		// Update existing user's profile image and name if changed
		firstName, lastName := validation.ParseFullName(googleUser.Name, googleUser.GivenName, googleUser.FamilyName)

		// Use helper function to build updates
		updates := validation.BuildGoogleUserUpdates(firstName, lastName, googleUser.Picture)
		updates["updated_at"] = time.Now()

		if err := s.UpdateUser(existingUser.UserID, updates); err != nil {
			return nil, err
		}

		// Return updated user
		return s.GetUserByID(existingUser.UserID)
	}

	// Create new user
	return s.CreateUser(googleUser)
}

func (s *UserService) GetUsersWithFilters(page, limit int, isTeacherFilter, search string) ([]models.User, int, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{})

	// Apply teacher filter
	if isTeacherFilter == "teacher" {
		query = query.Where("is_teacher = ?", true)
	} else if isTeacherFilter == "student" {
		query = query.Where("is_teacher = ?", false)
	}

	// Apply search filter (search in name and email)
	if search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Get total count for pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to count users")
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to fetch users")
	}

	return users, int(total), nil
}

func (s *UserService) UpdateTeacherStatus(userID string, isTeacher bool) error {
	updates := map[string]interface{}{
		"is_teacher": isTeacher,
		"updated_at": time.Now(),
	}

	return s.UpdateUser(userID, updates)
}

func (s *UserService) GetUserStatistics() (*types.UserStatistics, error) {
	stats := &types.UserStatistics{
		LastUpdated: time.Now(),
	}

	// Get total user count
	if err := s.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, errors.Wrap(err, "failed to count total users")
	}

	// Get teacher count
	if err := s.db.Model(&models.User{}).Where("is_teacher = ?", true).Count(&stats.TeacherCount).Error; err != nil {
		return nil, errors.Wrap(err, "failed to count teachers")
	}

	// Get student count
	stats.StudentCount = stats.TotalUsers - stats.TeacherCount

	// Get recent users (last 24 hours)
	yesterday := time.Now().Add(-24 * time.Hour)
	if err := s.db.Model(&models.User{}).
		Where("created_at >= ?", yesterday).
		Count(&stats.RecentUsers).Error; err != nil {
		return nil, errors.Wrap(err, "failed to count recent users")
	}

	// Get active users (updated in last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	if err := s.db.Model(&models.User{}).
		Where("updated_at >= ?", thirtyDaysAgo).
		Count(&stats.ActiveUsers).Error; err != nil {
		return nil, errors.Wrap(err, "failed to count active users")
	}

	return stats, nil
}

// Keep existing methods but update logic
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var userModel models.User
	if err := s.db.Where("email = ?", email).First(&userModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get user by email")
	}
	return &userModel, nil
}

func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	var userModel models.User
	if err := s.db.Where("user_id = ?", userID).First(&userModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get user by ID")
	}
	return &userModel, nil
}

func (s *UserService) UpdateUser(userID string, updates map[string]interface{}) error {
	result := s.db.Model(&models.User{}).Where("user_id = ?", userID).Updates(updates)
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to update user")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundError("User")
	}
	return nil
}

func (s *UserService) DeleteUser(userID string) error {
	result := s.db.Where("user_id = ?", userID).Delete(&models.User{})
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to delete user")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundError("User")
	}
	return nil
}

func (s *UserService) CreateUserDirect(user *models.User) error {
	if err := s.db.Create(user).Error; err != nil {
		return errors.Wrap(err, "failed to create user")
	}
	return nil
}

// Keep refresh token methods unchanged
func (s *UserService) StoreRefreshToken(userID, refreshToken string) error {
	updates := map[string]interface{}{
		"refresh_token":    refreshToken,
		"refresh_token_at": time.Now(),
	}
	return s.UpdateUser(userID, updates)
}

func (s *UserService) GetRefreshToken(userID string) (string, error) {
	var userModel models.User
	if err := s.db.Select("refresh_token, refresh_token_at").Where("user_id = ?", userID).First(&userModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.NewNotFoundError("User")
		}
		return "", errors.Wrap(err, "failed to get refresh token")
	}

	if userModel.RefreshTokenAt != nil && time.Since(*userModel.RefreshTokenAt) > 30*24*time.Hour {
		return "", errors.NewValidationError("refresh token too old", "refresh_token")
	}

	return userModel.RefreshToken, nil
}

func (s *UserService) RemoveRefreshToken(userID string) error {
	updates := map[string]interface{}{
		"refresh_token":    nil,
		"refresh_token_at": nil,
	}
	return s.UpdateUser(userID, updates)
}

func (s *UserService) IsRefreshTokenValid(userID, refreshToken string) (bool, error) {
	storedToken, err := s.GetRefreshToken(userID)
	if err != nil {
		return false, err
	}
	return storedToken == refreshToken && refreshToken != "", nil
}
