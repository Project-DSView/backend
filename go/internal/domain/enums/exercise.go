package enums

// ExerciseStatus represents the status of an exercise
type ExerciseStatus string

const (
	ExerciseStatusDraft     ExerciseStatus = "draft"
	ExerciseStatusPublished ExerciseStatus = "published"
	ExerciseStatusArchived  ExerciseStatus = "archived"
)

// IsValidExerciseStatus checks if the status is valid
func IsValidExerciseStatus(status string) bool {
	switch ExerciseStatus(status) {
	case ExerciseStatusDraft, ExerciseStatusPublished, ExerciseStatusArchived:
		return true
	default:
		return false
	}
}
