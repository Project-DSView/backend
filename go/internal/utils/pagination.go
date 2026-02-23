package utils

import (
	"strconv"
	"time"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
	Cursor string
}

// CursorPaginationParams holds cursor-based pagination parameters
type CursorPaginationParams struct {
	Limit  int
	Cursor string
	Order  string // "asc" or "desc"
}

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(page, limit int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}

	offset := (page - 1) * limit

	return PaginationParams{
		Limit:  limit,
		Offset: offset,
	}
}

// ParseCursorPaginationParams parses cursor-based pagination parameters
func ParseCursorPaginationParams(cursor string, limit int, order string) CursorPaginationParams {
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}
	if order != "asc" && order != "desc" {
		order = "desc" // Default order
	}

	return CursorPaginationParams{
		Limit:  limit,
		Cursor: cursor,
		Order:  order,
	}
}

// ParseTimestampCursor parses cursor as timestamp
func ParseTimestampCursor(cursor string) (time.Time, error) {
	if cursor == "" {
		return time.Time{}, nil
	}

	timestamp, err := strconv.ParseInt(cursor, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

// FormatTimestampCursor formats timestamp as cursor
func FormatTimestampCursor(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

// HasNextPage checks if there are more pages
func HasNextPage(results []interface{}, limit int) bool {
	return len(results) > limit
}

// GetNextCursor gets the next cursor for pagination
func GetNextCursor(results []interface{}, limit int, getTimestamp func(interface{}) time.Time) string {
	if len(results) <= limit {
		return ""
	}

	// Remove the extra item used for checking
	actualResults := results[:limit]
	if len(actualResults) == 0 {
		return ""
	}

	lastItem := actualResults[len(actualResults)-1]
	return FormatTimestampCursor(getTimestamp(lastItem))
}
