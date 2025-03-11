package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/katsukiniwa/kubernetes-sandbox/product/pkg/infrastructure/repository"
	"github.com/katsukiniwa/kubernetes-sandbox/product/pkg/usecase/command/create_order"
)

type PurchaseHandler interface {
	Purchase(w http.ResponseWriter, r *http.Request)
}

type purchaseHandler struct {
	pr repository.ProductRepository
	c  create_order.ICreateOrderCommand
}

func NewPurchaseHandler(pr repository.ProductRepository) PurchaseHandler {
	return &purchaseHandler{pr: pr, c: create_order.NewCreateOrderCommand()}
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
