package order

type OrderItem struct {
	OrderID       int64
	ProductID     int64
	Quantity      int64
	PurchasePrice int64
	TotalPrice    int64
}

func (item OrderItem) Subtotal() int64 {
	return int64(item.Quantity) * item.PurchasePrice
}
