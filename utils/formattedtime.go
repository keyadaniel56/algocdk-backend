package utils

import (
	"strings"
	"time"

	"database/sql/driver"
	"fmt"
)

type FormattedTime time.Time

// Define the format you want for API responses
const layout = "2006-01-02 15:04:05"

// MarshalJSON – controls how it appears in JSON responses
func (ft FormattedTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", time.Time(ft).Format(layout))
	return []byte(formatted), nil
}

// UnmarshalJSON – for reading from JSON
func (ft *FormattedTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(layout, s)
	if err != nil {
		return err
	}
	*ft = FormattedTime(t)
	return nil
}

// GORM scan interface
func (ft *FormattedTime) Scan(value interface{}) error {
	if t, ok := value.(time.Time); ok {
		*ft = FormattedTime(t)
		return nil
	}
	return fmt.Errorf("cannot scan value %v into FormattedTime", value)
}

// GORM value interface
func (ft FormattedTime) Value() (driver.Value, error) {
	return time.Time(ft), nil
}
