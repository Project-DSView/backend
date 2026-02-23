package week

import (
	"fmt"

	"gorm.io/gorm"
)

// FilterByWeek filters a slice of week-based entities by week number
func FilterByWeek[T WeekBasedEntity](items []T, week int) []T {
	var filtered []T
	for _, item := range items {
		if item.GetWeek() == week {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// GroupByWeek groups week-based entities by week number
func GroupByWeek[T WeekBasedEntity](items []T) map[int][]T {
	grouped := make(map[int][]T)
	for _, item := range items {
		week := item.GetWeek()
		grouped[week] = append(grouped[week], item)
	}
	return grouped
}

// GetWeekStatsFromDB retrieves week statistics from database
func GetWeekStatsFromDB(db *gorm.DB, tableName string, courseID string) (*WeekStats, error) {
	stats := &WeekStats{
		ByWeek: make(map[int]int64),
	}

	// Count total items
	var totalCount int64
	query := db.Table(tableName)
	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total items: %w", err)
	}
	stats.TotalCount = totalCount

	// Count by week
	var weekCounts []struct {
		Week  int   `json:"week"`
		Count int64 `json:"count"`
	}

	weekQuery := db.Table(tableName).
		Select("week, COUNT(*) as count").
		Where("week > 0") // Exclude week 0 (no specific week)

	if courseID != "" {
		weekQuery = weekQuery.Where("course_id = ?", courseID)
	}

	if err := weekQuery.Group("week").Scan(&weekCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get week counts: %w", err)
	}

	// Process week counts and find range
	minWeek, maxWeek := 0, 0
	for i, wc := range weekCounts {
		stats.ByWeek[wc.Week] = wc.Count

		if i == 0 {
			minWeek, maxWeek = wc.Week, wc.Week
		} else {
			if wc.Week < minWeek {
				minWeek = wc.Week
			}
			if wc.Week > maxWeek {
				maxWeek = wc.Week
			}
		}
	}

	stats.WeekRange = WeekRange{
		Min: minWeek,
		Max: maxWeek,
	}

	return stats, nil
}

// GetWeekGroupedDataFromDB retrieves data grouped by week from database
func GetWeekGroupedDataFromDB[T WeekBasedEntity](db *gorm.DB, entity T, filter WeekFilter) ([]WeekGroupedData[T], error) {
	var results []WeekGroupedData[T]

	// Get week counts
	var weekCounts []struct {
		Week  int   `json:"week"`
		Count int64 `json:"count"`
	}

	query := db.Table(entity.GetTableName()).
		Select("week, COUNT(*) as count").
		Where("week > 0") // Exclude week 0

	if filter.CourseID != "" {
		query = query.Where("course_id = ?", filter.CourseID)
	}

	if filter.Week != nil {
		query = query.Where("week = ?", *filter.Week)
	}

	if err := query.Group("week").Order("week ASC").Scan(&weekCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get week counts: %w", err)
	}

	// Get data for each week
	for _, wc := range weekCounts {
		var items []T

		itemQuery := db.Table(entity.GetTableName()).Where("week = ?", wc.Week)
		if filter.CourseID != "" {
			itemQuery = itemQuery.Where("course_id = ?", filter.CourseID)
		}

		if filter.Limit > 0 {
			itemQuery = itemQuery.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			itemQuery = itemQuery.Offset(filter.Offset)
		}

		if err := itemQuery.Find(&items).Error; err != nil {
			return nil, fmt.Errorf("failed to get items for week %d: %w", wc.Week, err)
		}

		results = append(results, WeekGroupedData[T]{
			Week:  wc.Week,
			Items: items,
			Count: int(wc.Count),
		})
	}

	return results, nil
}

// ValidateWeek validates week number
func ValidateWeek(week int) error {
	if week < 0 {
		return fmt.Errorf("week number cannot be negative: %d", week)
	}
	if week > 52 {
		return fmt.Errorf("week number cannot exceed 52: %d", week)
	}
	return nil
}

// GetWeekRange returns a slice of week numbers in the specified range
func GetWeekRange(startWeek, endWeek int) []int {
	if startWeek > endWeek {
		return []int{}
	}

	weeks := make([]int, endWeek-startWeek+1)
	for i := 0; i <= endWeek-startWeek; i++ {
		weeks[i] = startWeek + i
	}
	return weeks
}

// GetWeekContentFromDB retrieves content grouped by week with titles from database
func GetWeekContentFromDB[T WeekBasedEntity](db *gorm.DB, entity T, titleProvider WeekTitleProvider, filter WeekFilter) ([]WeekContent[T], error) {
	var results []WeekContent[T]

	// Get week counts
	var weekCounts []struct {
		Week  int   `json:"week"`
		Count int64 `json:"count"`
	}

	query := db.Table(entity.GetTableName()).
		Select("week, COUNT(*) as count").
		Where("week > 0") // Exclude week 0

	if filter.CourseID != "" {
		query = query.Where("course_id = ?", filter.CourseID)
	}

	if filter.Week != nil {
		query = query.Where("week = ?", *filter.Week)
	}

	if err := query.Group("week").Order("week ASC").Scan(&weekCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get week counts: %w", err)
	}

	// Get data for each week
	for _, wc := range weekCounts {
		var items []T

		itemQuery := db.Table(entity.GetTableName()).Where("week = ?", wc.Week)
		if filter.CourseID != "" {
			itemQuery = itemQuery.Where("course_id = ?", filter.CourseID)
		}

		if filter.Limit > 0 {
			itemQuery = itemQuery.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			itemQuery = itemQuery.Offset(filter.Offset)
		}

		if err := itemQuery.Find(&items).Error; err != nil {
			return nil, fmt.Errorf("failed to get items for week %d: %w", wc.Week, err)
		}

		// Get week title
		var title string
		if titleProvider != nil {
			title = titleProvider.GetWeekTitle(filter.CourseID, wc.Week)
		} else {
			title = fmt.Sprintf("สัปดาห์ที่ %d", wc.Week)
		}

		results = append(results, WeekContent[T]{
			WeekNumber: wc.Week,
			Title:      title,
			Items:      items,
			Count:      int(wc.Count),
		})
	}

	return results, nil
}

// GetPinnedItemsFromDB retrieves items without week (pinned items)
func GetPinnedItemsFromDB[T WeekBasedEntity](db *gorm.DB, entity T, filter WeekFilter) ([]T, error) {
	var items []T

	query := db.Table(entity.GetTableName()).Where("week IS NULL OR week = 0")

	if filter.CourseID != "" {
		query = query.Where("course_id = ?", filter.CourseID)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get pinned items: %w", err)
	}

	return items, nil
}
