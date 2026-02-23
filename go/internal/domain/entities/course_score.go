package models

import (
	"encoding/json"
	"time"
)

// CourseScore represents the domain entity for course scores
type CourseScore struct {
	UserID      string
	CourseID    string
	TotalScore  int
	LastUpdated time.Time
	CreatedAt   time.Time
}

// NewCourseScore creates a new CourseScore entity
func NewCourseScore(userID, courseID string, totalScore int) *CourseScore {
	return &CourseScore{
		UserID:      userID,
		CourseID:    courseID,
		TotalScore:  totalScore,
		LastUpdated: time.Now(),
		CreatedAt:   time.Now(),
	}
}

// UpdateScore updates the score
func (cs *CourseScore) UpdateScore(totalScore int) {
	cs.TotalScore = totalScore
	cs.LastUpdated = time.Now()
}

// IsPassing returns true if the student passed (score >= 60)
func (cs *CourseScore) IsPassing() bool {
	return cs.TotalScore >= 60
}

// ToJSON converts CourseScore to JSON string
func (cs *CourseScore) ToJSON() (string, error) {
	data, err := json.Marshal(cs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
