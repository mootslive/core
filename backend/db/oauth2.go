package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
)

type OAuth2Token oauth2.Token

// Scan implements the database/sql Scanner interface.
func (dst *OAuth2Token) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	switch src := src.(type) {
	case string:
		return json.Unmarshal([]byte(src), dst)
	case []byte:
		return json.Unmarshal(src, dst)
	}

	return fmt.Errorf("cannot scan %T", src)
}

// Value implements the database/sql/driver Valuer interface.
func (src OAuth2Token) Value() (driver.Value, error) {
	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	return data, err
}
