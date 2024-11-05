package objects

import (
	"github.com/google/uuid"
)

type Item struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4()"`
	ModelType string    `json:"model_type"`
	ModelID   int       `json:"model_id"`
	Category  string    `json:"category"`
	NodeID    int       `json:"node_id"`
}
