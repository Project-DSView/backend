package types

import (
	"database/sql/driver"
	"errors"
)

// JSONData is a custom type for handling JSON data in database
// It can handle both objects (map[string]interface{}) and arrays ([]interface{})
type JSONData []byte

// Value implements the driver.Valuer interface for database storage
func (j JSONData) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

// Scan implements the sql.Scanner interface for database reading
func (j *JSONData) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-[]byte value into JSONData")
	}

	*j = JSONData(bytes)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (j *JSONData) UnmarshalJSON(data []byte) error {
	*j = JSONData(data)
	return nil
}

// MarshalJSON implements json.Marshaler interface
func (j JSONData) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}
