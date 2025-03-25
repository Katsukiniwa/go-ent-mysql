package get_products

import (
	"context"
	"fmt"

	"github.com/katsukiniwa/go-ent-mysql/product/pkg/entity/product"
)

type IGetProductsQuery interface {
	Execute(ctx context.Context) ([]GetProductsDTO, error)
}

type getProductsQuery struct {
	pr product.IProductRepository
}

func NewGetProductsQuery(pr product.IProductRepository) IGetProductsQuery {
	return &getProductsQuery{pr: pr}
}

func (q *getProductsQuery) Execute(ctx context.Context) ([]GetProductsDTO, error) {
	products, err := q.pr.GetProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	var productsResponse []GetProductsDTO
	for _, v := range products {
		productsResponse = append(productsResponse, GetProductsDTO{ID: v.Id, Title: v.Title, Stock: int(v.Stock)})
	}

	return productsResponse, nil
}
