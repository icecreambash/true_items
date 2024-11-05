package objects

type Room struct {
	ID           int64    `json:"id" db:"id"`
	Rooms        int      `json:"rooms" db:"rooms"`
	AreaFull     float64  `json:"area_full" db:"area_full"`
	PriceFull    float64  `json:"price_full" db:"price_full"`
	PriceUnit    float64  `json:"price_unit" db:"price_unit"`
	NumberObject string   `json:"number_object" db:"number_object"`
	Status       string   `json:"status" db:"status"`
	Floor        int      `json:"floor" db:"floor"`
	Images       []string `json:"images" db:"images" gorm:"serializer:json"`
	Description  string   `json:"description" db:"description"`
	PlanType     string   `json:"plan_type" db:"plan_type"`
}

func (r Room) Empty() bool {
	if r.ID == 0 {
		return true
	}
	return false
}
