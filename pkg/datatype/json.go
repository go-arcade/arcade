package datatype

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/30 11:02
 * @file: json.go
 * @description: gorm json data type
 */

type JSON []byte

// Value implements the driver.Valuer interface
func (j *JSON) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}
	return []byte(*j), nil
}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = append((*j)[0:0], v...)
	default:
		return errors.New("unable to convert type to JSON")
	}
	return nil
}

// MarshalJSON implements the json.Marshal interface
func (j *JSON) MarshalJSON() ([]byte, error) {
	if j.IsNull() {
		return []byte("null"), nil
	}
	return *j, nil
}

// UnmarshalJSON implements the json.Unmarshal interface
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("null point exception")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Array returns the json as an array
func (j *JSON) Array() ([]interface{}, error) {
	if j.IsNull() {
		return nil, nil
	}
	var arr []interface{}
	if err := json.Unmarshal(*j, &arr); err != nil {
		return nil, err
	}
	return arr, nil
}

// Map returns the json as a map
func (j *JSON) Map() (map[interface{}]interface{}, error) {
	if j.IsNull() {
		return nil, nil
	}
	var m map[interface{}]interface{}
	if err := json.Unmarshal(*j, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Field returns the value of a field in the json
func (j *JSON) Field(key string) (interface{}, error) {
	if j.IsNull() {
		return nil, nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(*j, &m); err != nil {
		return nil, err
	}
	val, exists := m[key]
	if !exists {
		return nil, errors.New("field does not exist")
	}
	return val, nil
}

// String returns the json as a string
func (j *JSON) String() (string, error) {
	if j.IsNull() {
		return "", nil
	}
	return string(*j), nil
}

// IsNull returns true if the json is null
func (j *JSON) IsNull() bool {
	return len(*j) == 0 || string(*j) == "null"
}

// Equals returns true if the json is equal to the given json
func (j *JSON) Equals(json JSON) bool {
	return bytes.Equal(*j, json)
}
