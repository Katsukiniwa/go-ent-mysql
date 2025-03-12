package product

import "time"

type ProductPriceHistory struct {
	ProductID int
	Price     int64
	StartedAt time.Time
	EndedAt   time.Time
}
