package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/katsukiniwa/kubernetes-sandbox/product/pkg/infrastructure/repository"
	"github.com/katsukiniwa/kubernetes-sandbox/product/pkg/usecase/query/get_products"
)

type GetProductsHandler interface {
	GetProducts(w http.ResponseWriter, r *http.Request)
}

type getProductsHandler struct {
	pr repository.ProductRepository
	q  get_products.IGetProductsQuery
}

func NewGetProductsHandler(pr repository.ProductRepository) GetProductsHandler {
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
