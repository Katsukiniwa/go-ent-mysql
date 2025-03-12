package order

import (
	"time"

	"github.com/google/uuid"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/entity/customer"
)

type Order struct {
	ID         uuid.UUID
	CustomerID int
	Items      []OrderItem
	OrderedAt  time.Time
	TotalPrice int64
	Status     Status
}

type Status int

const (
	Complete Status = iota
	Cancel
	Shipping
	Shipped
)

func NewOrder(customer customer.Customer, items []OrderItem) Order {
	var total int64
	for _, item := range items {
		total += item.Subtotal()
	}

	return Order{
		ID:         uuid.New(),
		CustomerID: customer.ID,
		Items:      items,
		TotalPrice: total,
		OrderedAt:  time.Now(),
	}
}
