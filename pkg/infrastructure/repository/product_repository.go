package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/katsukiniwa/kubernetes-sandbox/product/ent"
	"github.com/katsukiniwa/kubernetes-sandbox/product/pkg/entity/product"
)

type ProductRepository interface {
	GetProducts(ctx context.Context) ([]product.Product, error)
	GetByID(ctx context.Context, id int) (*product.Product, error)
	InsertProduct(ctx context.Context, title string) error
	UpdateProduct(ctx context.Context, product *product.Product) error
	DeleteProduct(ctx context.Context, id int) error
}

type productRepository struct {
	client *ent.Client
}

func NewProductRepository(client *ent.Client) ProductRepository {
	return &productRepository{client: client}
}

func (pr *productRepository) GetProducts(ctx context.Context) ([]product.Product, error) {
	products, err := pr.client.Product.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying products: %w", err)
	}

	log.Println("products: ", products)

	var result []product.Product

	for _, v := range products {
		result = append(result, product.Product{Id: v.ID, Title: v.Title, Stock: int64(v.Stock)})
	}

	return result, nil
}

func (pr *productRepository) GetByID(ctx context.Context, id int) (*product.Product, error) {
	p, err := pr.client.Product.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed updating product: %w", err)
	}
	return &product.Product{Id: p.ID, Title: p.Title, Stock: int64(p.Stock)}, nil
}

func (pr *productRepository) UpdateProduct(ctx context.Context, p *product.Product) error {
	_, err := pr.client.Product.UpdateOneID(p.Id).SetTitle(p.Title).SetStock(int(p.Stock)).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed updating product: %w", err)
	}
	return nil
}

func (pr *productRepository) DeleteProduct(ctx context.Context, id int) error {
	err := pr.client.Product.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed updating product: %w", err)
	}
	return nil
}

func (pr *productRepository) InsertProduct(ctx context.Context, title string) error {
	_, err := pr.client.Product.Create().SetTitle(title).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed updating product: %w", err)
	}
	return nil
}
