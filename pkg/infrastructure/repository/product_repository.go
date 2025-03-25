package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/katsukiniwa/go-ent-mysql/product/ent"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/entity/product"
)

type ProductRepository struct {
	client *ent.Client
}

func NewProductRepository(client *ent.Client) product.IProductRepository {
	return &ProductRepository{client: client}
}

func (pr *ProductRepository) GetProducts(ctx context.Context) ([]product.Product, error) {
	products, err := pr.client.Product.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying products: %w", err)
	}

	log.Println("products: ", products)

	result := make([]product.Product, 0, len(products))

	for _, v := range products {
		result = append(result, product.Product{Id: v.ID, Title: v.Title, Stock: int64(v.Stock)})
	}

	return result, nil
}

func (pr *ProductRepository) GetByID(ctx context.Context, id int) (*product.Product, error) {
	p, err := pr.client.Product.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed updating product: %w", err)
	}

	return &product.Product{Id: p.ID, Title: p.Title, Stock: int64(p.Stock)}, nil
}

func (pr *ProductRepository) UpdateProduct(ctx context.Context, p *product.Product) error {
	_, err := pr.client.Product.UpdateOneID(p.Id).SetTitle(p.Title).SetStock(int(p.Stock)).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed updating product: %w", err)
	}

	return nil
}

func (pr *ProductRepository) DeleteProduct(ctx context.Context, id int) error {
	err := pr.client.Product.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed updating product: %w", err)
	}

	return nil
}

func (pr *ProductRepository) InsertProduct(ctx context.Context, title string) error {
	_, err := pr.client.Product.Create().SetTitle(title).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed updating product: %w", err)
	}

	return nil
}
