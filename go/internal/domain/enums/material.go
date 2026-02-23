package enums

// MaterialType represents the type of course material
type MaterialType string

const (
	MaterialTypeAnnouncement MaterialType = "announcement"   // ประกาศ (Announcements)
	MaterialTypeDocument     MaterialType = "document"      // ไฟล์เอกสาร (PDF, DOC, etc.)
	MaterialTypeVideo        MaterialType = "video"         // วิดีโอ (YouTube, Vimeo, etc.)
	MaterialTypeCodeExercise MaterialType = "code_exercise" // แบบฝึกหัดโค้ด (Code exercises)
	MaterialTypePDFExercise  MaterialType = "pdf_exercise"  // แบบฝึกหัด PDF (PDF exercises)
)
