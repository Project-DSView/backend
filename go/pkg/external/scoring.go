package external

import (
	"encoding/json"
	"reflect"
)

func CompareJSON(expected, actual interface{}) bool {
	// Convert both to JSON strings for comparison
	expectedJSON, err1 := json.Marshal(expected)
	actualJSON, err2 := json.Marshal(actual)

	if err1 != nil || err2 != nil {
		return reflect.DeepEqual(expected, actual)
	}

	// Normalize by unmarshaling back to interface{}
	var expectedNormalized, actualNormalized interface{}
	json.Unmarshal(expectedJSON, &expectedNormalized)
	json.Unmarshal(actualJSON, &actualNormalized)

	return reflect.DeepEqual(expectedNormalized, actualNormalized)
}

func ScoreFromCounts(totalPoints, passed, total int) int {
	if total <= 0 {
		return 0
	}
	if totalPoints < 0 {
		totalPoints = 0
	}
	return (totalPoints * passed) / total
}
