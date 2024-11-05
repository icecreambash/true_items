package objects

import "time"

type Complex struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	DeletedAt time.Time `json:"deleted_at" db:"deleted_at"`
}

func (a Complex) GetID() int64 {
	return a.ID
}
