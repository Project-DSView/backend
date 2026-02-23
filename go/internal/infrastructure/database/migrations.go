package database

import (
	"fmt"

	"gorm.io/gorm"
)

// createCourseScoreIndexes creates indexes for course score table
func createCourseScoreIndexes(db *gorm.DB) error {
	indexes := []string{
		// Composite index for user and course lookups
		"CREATE INDEX IF NOT EXISTS idx_student_course_scores_user_course ON student_course_scores(user_id, course_id)",

		// Index for course leaderboard queries
		"CREATE INDEX IF NOT EXISTS idx_student_course_scores_course_score ON student_course_scores(course_id, total_score DESC, last_updated ASC)",

		// Index for student statistics
		"CREATE INDEX IF NOT EXISTS idx_student_course_scores_user ON student_course_scores(user_id)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", indexSQL, err)
		}
	}

	return nil
}

// createStudentProgressIndexes creates indexes for student progress table
func createStudentProgressIndexes(db *gorm.DB) error {
	indexes := []string{
		// Composite index for user and material lookups
		"CREATE INDEX IF NOT EXISTS idx_student_progress_user_material ON student_progress(user_id, material_id)",

		// Index for course progress queries
		"CREATE INDEX IF NOT EXISTS idx_student_progress_user ON student_progress(user_id)",

		// Index for material progress queries
		"CREATE INDEX IF NOT EXISTS idx_student_progress_material ON student_progress(material_id)",

		// Index for status-based queries
		"CREATE INDEX IF NOT EXISTS idx_student_progress_status ON student_progress(status)",

		// Index for score-based queries
		"CREATE INDEX IF NOT EXISTS idx_student_progress_score ON student_progress(score)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", indexSQL, err)
		}
	}

	return nil
}

// createSubmissionIndexes creates indexes for submission table
func createSubmissionIndexes(db *gorm.DB) error {
	indexes := []string{
		// Composite index for user and material lookups
		"CREATE INDEX IF NOT EXISTS idx_submissions_user_material ON submissions(user_id, material_id)",

		// Index for status-based queries
		"CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status)",

		// Index for submitted_at ordering
		"CREATE INDEX IF NOT EXISTS idx_submissions_submitted_at ON submissions(submitted_at)",

		// Index for score-based queries
		"CREATE INDEX IF NOT EXISTS idx_submissions_total_score ON submissions(total_score)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", indexSQL, err)
		}
	}

	return nil
}

// AnalyzeTables runs ANALYZE on tables to update statistics
func AnalyzeTables(db *gorm.DB) error {
	tables := []string{
		"student_course_scores",
		"student_progress",
		"submissions",
		"users",
		"courses",
		"course_materials",
		"test_cases",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("ANALYZE %s", table)).Error; err != nil {
			return fmt.Errorf("failed to analyze table %s: %w", table, err)
		}
	}

	return nil
}
