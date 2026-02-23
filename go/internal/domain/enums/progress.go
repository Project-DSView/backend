package enums

// ProgressStatus represents the status of student progress
type ProgressStatus string

const (
	ProgressNotStarted      ProgressStatus = "not_started"
	ProgressInProgress      ProgressStatus = "in_progress"
	ProgressWaitingReview   ProgressStatus = "waiting_review"
	ProgressWaitingApproval ProgressStatus = "waiting_approval"
	ProgressCompleted       ProgressStatus = "completed"
)
