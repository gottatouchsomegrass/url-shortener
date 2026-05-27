//Package models inclue all the models used
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
	//"github.com/google/uuid"
)

type URL struct {
	ID			int64		`json:"id" db:"id"`//go for uuid later
	LongURL		string		`json:"long_url" db:"long_url" validate:"required,url"`
	ShortURL	string		`json:"short_url" db:"short_url" validate:"required,shortcode"`
	Expiry		*time.Time	`json:"expiry,omitempty" db:"expiry"`
	Clicks		int64		`json:"clicks" db:"clicks"`
	CreatedAt	time.Time	`json:"created_at" db:"created_at"`
}

func (u URL) Value() (driver.Value, error) {
	return json.Marshal(u)
}

func (u *URL) Scan(value any) error {
	j, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(j, &u)
}
