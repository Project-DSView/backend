package week

// WeekBasedEntity defines the interface for entities that have week functionality
type WeekBasedEntity interface {
	// GetWeek returns the week number for this entity
	GetWeek() int

	// SetWeek sets the week number for this entity
	SetWeek(week int)

	// GetTableName returns the database table name for this entity
	GetTableName() string

	// GetCourseID returns the course ID associated with this entity (if applicable)
	GetCourseID() string
}

// WeekTitleProvider defines the interface for entities that can provide week titles
type WeekTitleProvider interface {
	// GetWeekTitle returns the title for a specific week
	GetWeekTitle(courseID string, weekNumber int) string
}

// WeekFilter represents filtering criteria for week-based queries
type WeekFilter struct {
	CourseID string
	Week     *int // nil means no week filter
	Limit    int
	Offset   int
}

// WeekStats represents statistics for week-based data
type WeekStats struct {
	TotalCount int64         `json:"total_count"`
	ByWeek     map[int]int64 `json:"by_week"`
	WeekRange  WeekRange     `json:"week_range"`
}

// WeekRange represents the range of weeks in the data
type WeekRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// WeekGroupedData represents data grouped by week
type WeekGroupedData[T WeekBasedEntity] struct {
	Week  int `json:"week"`
	Items []T `json:"items"`
	Count int `json:"count"`
}

// WeekContent represents content for a specific week with title
type WeekContent[T WeekBasedEntity] struct {
	WeekNumber int    `json:"week_number"`
	Title      string `json:"title"`
	Items      []T    `json:"items"`
	Count      int    `json:"count"`
}
