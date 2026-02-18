package utils

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// NullString is a helper that wraps sql.NullString and marshals null to empty string
// and implements sql.Scanner and driver.Valuer so it works with sqlx.
type NullString struct {
	sql.NullString
}

// NewNullString creates a NullString from a plain string. Empty string -> invalid (null)
func NewNullString(s string) NullString {
	if s == "" {
		return NullString{sql.NullString{String: "", Valid: false}}
	}
	return NullString{sql.NullString{String: s, Valid: true}}
}

// MarshalJSON returns the string value or an empty string when null.
func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return json.Marshal("")
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON supports both JSON string and null. Null -> invalid.
func (ns *NullString) UnmarshalJSON(b []byte) error {
	// if JSON null
	if string(b) == "null" {
		ns.String = ""
		ns.Valid = false
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	ns.String = s
	ns.Valid = true
	return nil
}

// Scan implements sql.Scanner
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}
	switch v := value.(type) {
	case string:
		ns.String = v
	case []byte:
		ns.String = string(v)
	default:
		ns.String = fmt.Sprintf("%v", v)
	}
	ns.Valid = true
	return nil
}

// Value implements driver.Valuer
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

// String returns the inner string or empty when null
func (ns NullString) StringValue() string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

// NullInt64 is a helper that wraps sql.NullInt64 and marshals null to JSON null
// and implements sql.Scanner and driver.Valuer so it works with sqlx.
type NullInt64 struct {
	sql.NullInt64
}

// NewNullInt64 creates a NullInt64 from an int64. Zero value -> valid.
func NewNullInt64(i int64) NullInt64 {
	return NullInt64{sql.NullInt64{Int64: i, Valid: true}}
}

// MarshalJSON returns the integer value or null when invalid.
func (ni NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(ni.Int64)
}

// UnmarshalJSON supports both JSON number and null. Null -> invalid.
func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		ni.Int64 = 0
		ni.Valid = false
		return nil
	}
	var v int64
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	ni.Int64 = v
	ni.Valid = true
	return nil
}

// Scan implements sql.Scanner
func (ni *NullInt64) Scan(value interface{}) error {
	if value == nil {
		ni.Int64, ni.Valid = 0, false
		return nil
	}
	switch v := value.(type) {
	case int64:
		ni.Int64 = v
	case int32:
		ni.Int64 = int64(v)
	case int:
		ni.Int64 = int64(v)
	case []byte:
		// parse bytes as integer
		var i int64
		if err := json.Unmarshal(v, &i); err != nil {
			// fallback to string parsing
			s := string(v)
			var parsed int64
			_, err := fmt.Sscan(s, &parsed)
			if err != nil {
				return err
			}
			ni.Int64 = parsed
		} else {
			ni.Int64 = i
		}
	default:
		ni.Int64 = 0
	}
	ni.Valid = true
	return nil
}

// Value implements driver.Valuer
func (ni NullInt64) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return ni.Int64, nil
}

// Int64Value returns the inner int64 or zero when null
func (ni NullInt64) Int64Value() int64 {
	if !ni.Valid {
		return 0
	}
	return ni.Int64
}
