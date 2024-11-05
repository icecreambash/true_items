package db_models

import "github.com/google/uuid"

type Tenant struct {
	ID uuid.UUID `json:"id" db:"id"`
}
