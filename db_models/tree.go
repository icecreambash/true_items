package db_models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strconv"
)

type CustomData struct {
	IntValue  *int       `json:"int_value,omitempty"`
	UUIDValue *uuid.UUID `json:"uuid_value,omitempty"`
}

func (c *CustomData) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		// Try unmarshalling as JSON
		if err := json.Unmarshal(v, c); err == nil {
			return nil
		}
		// If that fails, try to handle it as a string (could be an int or uuid string)
		return c.handleString(string(v))
	case string:
		return c.handleString(v)
	case int64:
		c.IntValue = new(int)
		*c.IntValue = int(v)
		return nil
	default:
		return fmt.Errorf("failed to scan data: %v", value)
	}
}

func (c *CustomData) handleString(value string) error {
	if intValue, err := strconv.Atoi(value); err == nil {
		c.IntValue = new(int)
		*c.IntValue = intValue
		return nil
	}
	if parsedUUID, err := uuid.Parse(value); err == nil {
		c.UUIDValue = &parsedUUID
		return nil
	}
	return fmt.Errorf("invalid string value for model_id: %v", value)
}

func (c CustomData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type Tree struct {
	ID        int        `db:"id" json:"ID"`
	ModelType string     `db:"model_type" json:"model_type"`
	ModelID   CustomData `db:"model_id" json:"model_id" gorm:"type:json"`
	ParentID  int        `db:"parent_id" json:"parent_id"`
}
