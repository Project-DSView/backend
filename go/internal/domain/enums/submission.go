package enums

// SubmissionStatus represents the status of a submission
type SubmissionStatus string

const (
	SubmissionPending   SubmissionStatus = "pending"
	SubmissionRunning   SubmissionStatus = "running"
	SubmissionError     SubmissionStatus = "error"
	SubmissionCompleted SubmissionStatus = "completed"
)

// VerificationStatus represents the verification status
type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "pending"
	VerificationApproved VerificationStatus = "approved"
	VerificationRejected VerificationStatus = "rejected"
)

// SubmissionResultStatus represents the result status of a submission
type SubmissionResultStatus string

const (
	SubmissionResultPassed SubmissionResultStatus = "passed"
	SubmissionResultFailed SubmissionResultStatus = "failed"
)
