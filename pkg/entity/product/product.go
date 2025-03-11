package product

import "fmt"

type Product struct {
	Id         int
	Title      string
	Stock      int64
	SaleStatus SaleStatus
}

// 販売ステータス
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
