package get_products

import (
	"context"

	"github.com/katsukiniwa/kubernetes-sandbox/product/pkg/infrastructure/repository"
)

type IGetProductsQuery interface {
	Execute(ctx context.Context) ([]GetProductsDTO, error)
}

type getProductsQuery struct {
	pr repository.ProductRepository
}

func NewGetProductsQuery(pr repository.ProductRepository) IGetProductsQuery {
	return &getProductsQuery{pr: pr}
}

func (q *getProductsQuery) Execute(ctx context.Context) ([]GetProductsDTO, error) {
	products, err := q.pr.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	var productsResponse []GetProductsDTO
	for _, v := range products {
		productsResponse = append(productsResponse, GetProductsDTO{ID: v.Id, Title: v.Title, Stock: int(v.Stock)})
	}

	return productsResponse, nil
}
