package objects

type Building struct {
	ID           int64  `json:"id" db:"id"`
	NumberObject string `json:"number_object" db:"number_object"`
}

func (a Building) GetID() int64 {
	return a.ID
}
