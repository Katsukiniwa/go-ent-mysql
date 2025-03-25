package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/katsukiniwa/go-ent-mysql/product/pkg/entity/product"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/usecase/command/create_order"
)

type PurchaseHandler interface {
	Purchase(w http.ResponseWriter, r *http.Request)
}

type purchaseHandler struct {
	pr product.IProductRepository
	c  create_order.ICreateOrderCommand
}

func NewPurchaseHandler(pr product.IProductRepository) PurchaseHandler {
	return &purchaseHandler{pr: pr, c: create_order.NewCreateOrderCommand(pr)}
}

func (pc *purchaseHandler) Purchase(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	decoder := json.NewDecoder(r.Body)

	var purchaseRequest create_order.CreateOrderDTO

	if err := decoder.Decode(&purchaseRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)

		return
	}

	err := pc.c.Execute(ctx, purchaseRequest)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	output, _ := json.MarshalIndent("Success", "", "\t\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
