package create_order

type CreateOrderDTO struct {
	CustomerID int         `json:"customerId"`
	Items      []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID int `json:"productId"`
	Quantity  int `json:"quantity"`
}
