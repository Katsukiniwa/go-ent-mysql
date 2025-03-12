package product

import (
	"context"
)

type IProductRepository interface {
	GetProducts(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id int) (*Product, error)
	InsertProduct(ctx context.Context, title string) error
	UpdateProduct(ctx context.Context, product *Product) error
	DeleteProduct(ctx context.Context, id int) error
}
