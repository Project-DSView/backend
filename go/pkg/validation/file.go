package validation

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize = 1024 * 1024 // 1MB
	MaxLines    = 10000       // จำกัดจำนวนบรรทัด
)

type FileValidationError struct {
	Field   string
	Message string
}

func (e *FileValidationError) Error() string {
	return e.Message
}

// ValidatePythonFile ตรวจสอบไฟล์ Python
func ValidatePythonFile(filename string, content []byte) error {
	// ตรวจสอบ extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".py" {
		return &FileValidationError{
			Field:   "file_type",
			Message: "Only Python (.py) files are allowed",
		}
	}

	// ตรวจสอบขนาดไฟล์
	if len(content) > MaxFileSize {
		return &FileValidationError{
			Field:   "file_size",
			Message: fmt.Sprintf("File size exceeds maximum limit of %d bytes", MaxFileSize),
		}
	}

	// ตรวจสอบว่าไฟล์ไม่ว่าง
	if len(content) == 0 {
		return &FileValidationError{
			Field:   "file_content",
			Message: "File cannot be empty",
		}
	}

	// ตรวจสอบจำนวนบรรทัด
	lines := strings.Split(string(content), "\n")
	if len(lines) > MaxLines {
		return &FileValidationError{
			Field:   "file_lines",
			Message: fmt.Sprintf("File has too many lines (max: %d)", MaxLines),
		}
	}

	// ตรวจสอบ basic security - ห้ามใช้คำสำคัญอันตราย
	contentStr := strings.ToLower(string(content))
	dangerousKeywords := []string{
		"__import__",
		"eval(",
		"exec(",
		"compile(",
		"open(",
		"file(",
		"input(",
		"raw_input(",
	}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(contentStr, keyword) {
			return &FileValidationError{
				Field:   "security",
				Message: fmt.Sprintf("File contains restricted keyword: %s", keyword),
			}
		}
	}

	return nil
}

// SanitizeFilename ทำความสะอาดชื่อไฟล์
func SanitizeFilename(filename string) string {
	// ลบ path และเอาแต่ชื่อไฟล์
	filename = filepath.Base(filename)

	// แทนที่อักขระที่ไม่ปลอดภัย
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "..", ".")

	return filename
}

// ValidateCodeContent ตรวจสอบเนื้อหาโค้ดโดยตรง
func ValidateCodeContent(code string) error {
	if strings.TrimSpace(code) == "" {
		return &FileValidationError{
			Field:   "code",
			Message: "Code cannot be empty",
		}
	}

	if len(code) > MaxFileSize {
		return &FileValidationError{
			Field:   "code_size",
			Message: "Code content is too large",
		}
	}

	return nil
}
