package product

import (
	"fmt"
	"time"
)

type Product struct {
	Id             int
	Title          string
	Stock          int64
	SaleStatus     SaleStatus
	PriceHistories []ProductPriceHistory
}

// 販売ステータス.
type SaleStatus int

const (
	SaleStatusOnSale SaleStatus = iota
	SaleStatusSoldOut
)

func (p *Product) DecreaseStock(amount int64) error {
	if p.Stock < amount {
		return fmt.Errorf("stock is not enough: %d", p.Stock)
	}

	p.Stock -= amount
	if p.Stock == 0 {
		p.SaleStatus = SaleStatusSoldOut
	}

	return nil
}

func (p *Product) CurrentPrice(t time.Time) int64 {
	result := int64(0)

	for _, v := range p.PriceHistories {
		if v.StartedAt.Before(t) && v.EndedAt.After(t) {
			result = v.Price
		}

		if v.EndedAt.IsZero() && v.StartedAt.Before(t) {
			result = v.Price
		}
	}

	return result
}
