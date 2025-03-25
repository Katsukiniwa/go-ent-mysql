package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/katsukiniwa/go-ent-mysql/product/pkg/entity/product"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/usecase/query/get_products"
)

type GetProductsHandler interface {
	GetProducts(w http.ResponseWriter, r *http.Request)
}

type getProductsHandler struct {
	pr product.IProductRepository
	q  get_products.IGetProductsQuery
}

func NewGetProductsHandler(pr product.IProductRepository) GetProductsHandler {
	return &getProductsHandler{pr: pr, q: get_products.NewGetProductsQuery(pr)}
}

func (pc *getProductsHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	productsResponse, err := pc.q.Execute(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	output, _ := json.MarshalIndent(productsResponse, "", "\t\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
