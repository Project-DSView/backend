package interfaces

import (
	"context"

	models "github.com/Project-DSView/backend/go/internal/domain/entities"
	"github.com/Project-DSView/backend/go/internal/types"
)

// UserServiceInterface defines the interface for user operations
type UserServiceInterface interface {
	CreateUser(googleUser *types.GoogleUser) (*models.User, error)
	CreateOrUpdateUser(googleUser *types.GoogleUser) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(userID string) (*models.User, error)
	GetUsersWithFilters(page, limit int, isTeacherFilter, search string) ([]models.User, int, error)
	UpdateUser(userID string, updates map[string]interface{}) error
	DeleteUser(userID string) error
	UpdateTeacherStatus(userID string, isTeacher bool) error
	GetUserStatistics() (*types.UserStatistics, error)
	StoreRefreshToken(userID, refreshToken string) error
	GetRefreshToken(userID string) (string, error)
	RemoveRefreshToken(userID string) error
	IsRefreshTokenValid(userID, refreshToken string) (bool, error)
}

// CourseMaterialServiceInterface defines the interface for course material operations
type CourseMaterialServiceInterface interface {
	CreateCourseMaterial(material *models.CourseMaterial) error
	GetCourseMaterialByID(materialID string) (map[string]interface{}, error)
	GetCourseMaterialsWithFilters(page, limit int, courseID, materialType, search, createdBy string, isTeacher bool) ([]models.CourseMaterial, int, error)
	UpdateCourseMaterial(materialID string, userID string, updates map[string]interface{}) error
	DeleteCourseMaterial(materialID string, userID string) error
	GetCourseMaterialStatistics() (map[string]interface{}, error)
	CanUserModifyCourseMaterial(userID, materialID string, isTeacher bool) (bool, error)
	GetCourseMaterialsByCourse(courseID string, week *int, materialType *string, limit, offset int) ([]map[string]interface{}, int64, error)
	CreateCodeExercise(codeExercise *models.CodeExercise, testCases []models.TestCase) error
	GetTestCases(materialID string) ([]models.TestCase, error)
	AddTestCase(materialID string, testCase *models.TestCase) error
	UpdateTestCase(testCaseID string, userID string, updates map[string]interface{}) error
	DeleteTestCase(testCaseID string, userID string) error
	GetCourseMaterialWithTestCases(materialID string) (*models.CourseMaterial, []models.TestCase, error)
}

// CourseServiceInterface defines the interface for course operations
type CourseServiceInterface interface {
	CreateCourse(course *models.Course) error
	GetCourseByID(courseID string) (*models.Course, error)
	GetCoursesWithFilters(page, limit int, status, search, createdBy string, isTeacher bool) ([]models.Course, int, error)
	UpdateCourse(courseID string, updates map[string]interface{}) error
	DeleteCourse(courseID string) error
	GetCourseStatistics() (map[string]interface{}, error)
	CanUserModifyCourse(userID, courseID string, isTeacher bool) (bool, error)
}

// TestCaseServiceInterface defines the interface for test case operations
type TestCaseServiceInterface interface {
	CreateTestCase(testCase *models.TestCase) error
	GetTestCaseByID(testCaseID string) (*models.TestCase, error)
	GetTestCasesByMaterialID(materialID string) ([]models.TestCase, error)
	UpdateTestCase(testCaseID string, updates map[string]interface{}) error
	DeleteTestCase(testCaseID string) error
	GetTestCaseStatistics() (map[string]interface{}, error)
}

// SubmissionServiceInterface defines the interface for submission operations
type SubmissionServiceInterface interface {
	SubmitMaterialExercise(userID, materialID, code string) (*types.SubmitResult, error)
	GetSubmissionByID(submissionID string) (*models.Submission, error)
	GetSubmissionsByMaterial(materialID string, page, limit int) ([]models.Submission, int, error)
	GetSubmissionsByUser(userID string, page, limit int) ([]models.Submission, int, error)
	GetSubmissionStatistics() (map[string]interface{}, error)
}

// ProgressServiceInterface defines the interface for progress operations
type ProgressServiceInterface interface {
	GetUserProgress(userID string) (*models.StudentProgress, error)
	UpdateProgress(userID, materialID string, score int) error
	GetProgressStatistics() (map[string]interface{}, error)
}

// DraftServiceInterface defines the interface for draft operations
type DraftServiceInterface interface {
	SaveDraft(userID, materialID, code, fileName string, fileSize int64) (*models.ExerciseDraft, error)
	GetDraft(userID, materialID string) (*models.ExerciseDraft, error)
	DeleteDraft(userID, materialID string) error
	GetUserDrafts(userID string) ([]models.ExerciseDraft, error)
}

// QueueServiceInterface defines the interface for queue operations
type QueueServiceInterface interface {
	EnqueueJob(job *models.QueueJob) error
	ProcessJob(jobID string) error
	GetJobStatus(jobID string) (*models.QueueJob, error)
	StartQueueConsumer(ctx context.Context) error
	StopQueueConsumer() error
}

// StorageServiceInterface defines the interface for storage operations
type StorageServiceInterface interface {
	UploadFile(bucket, key string, data []byte) error
	DownloadFile(bucket, key string) ([]byte, error)
	DeleteFile(bucket, key string) error
	GetFileURL(bucket, key string) (string, error)
	HealthCheck() error
}
