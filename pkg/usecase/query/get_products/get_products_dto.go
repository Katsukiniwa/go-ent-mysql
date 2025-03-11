package get_products

type GetProductsDTO struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Stock int    `json:"stock"`
}
