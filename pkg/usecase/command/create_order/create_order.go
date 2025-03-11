package create_order

import (
	"context"

	"github.com/katsukiniwa/kubernetes-sandbox/product/pkg/infrastructure/repository"
)

type ICreateOrderCommand interface {
	Execute(ctx context.Context, params CreateOrderDTO) error
}

type CreateOrderCommand struct {
	productRepository repository.ProductRepository
}

func NewCreateOrderCommand() ICreateOrderCommand {
	return &CreateOrderCommand{}
}

func (c *CreateOrderCommand) Execute(ctx context.Context, params CreateOrderDTO) error {
	for _, purchaseRequest := range params.Items {
		product, err := c.productRepository.GetByID(ctx, purchaseRequest.ProductID)
		if err != nil {
			return err
		}

		err = product.DecreaseStock(int64(purchaseRequest.Quantity))
		if err != nil {
			return err
		}

		err = c.productRepository.UpdateProduct(ctx, product)
		if err != nil {
			return err
		}
	}

	return nil
}
