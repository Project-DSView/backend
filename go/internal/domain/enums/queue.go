package enums

// QueueStatus represents the status of a queue job
type QueueStatus string

const (
	QueueStatusPending    QueueStatus = "pending"
	QueueStatusProcessing QueueStatus = "processing"
	QueueStatusCompleted  QueueStatus = "completed"
	QueueStatusFailed     QueueStatus = "failed"
	QueueStatusCancelled  QueueStatus = "cancelled"
)

// QueueType represents the type of a queue job
type QueueType string

const (
	QueueTypeCodeExecution QueueType = "code_execution"
	QueueTypeReview        QueueType = "review" // Renamed from code_review for unified handling
	QueueTypeFileProcessing QueueType = "file_processing"
)
