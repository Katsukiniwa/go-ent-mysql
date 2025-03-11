package create_order

type CreateOrderDTO struct {
	CustomerID int
	Items      []OrderItem
}

type OrderItem struct {
	ProductID int
	Quantity  int
}
