package create_order

import (
	"context"
	"fmt"

	"github.com/katsukiniwa/go-ent-mysql/product/pkg/entity/product"
)

type ICreateOrderCommand interface {
	Execute(ctx context.Context, params CreateOrderDTO) error
}

type CreateOrderCommand struct {
	pr product.IProductRepository
}

func NewCreateOrderCommand(pr product.IProductRepository) ICreateOrderCommand {
	return &CreateOrderCommand{pr: pr}
}

func (c *CreateOrderCommand) Execute(ctx context.Context, params CreateOrderDTO) error {
	for _, purchaseRequest := range params.Items {
		product, err := c.pr.GetByID(ctx, purchaseRequest.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get product: %w", err)
		}

		err = product.DecreaseStock(int64(purchaseRequest.Quantity))
		if err != nil {
			return fmt.Errorf("failed to decrease stock of products: %w", err)
		}

		err = c.pr.UpdateProduct(ctx, product)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}
	}

	return nil
}
