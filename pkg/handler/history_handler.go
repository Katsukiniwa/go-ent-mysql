package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/katsukiniwa/go-ent-mysql/product/pkg/handler/dto"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/infrastructure/repository"
)

type HistoryController interface {
	GetHistories(w http.ResponseWriter, r *http.Request)
	PostHistory(w http.ResponseWriter, r *http.Request)
	// PutHistory(w http.ResponseWriter, r *http.Request)
	// DeleteHistory(w http.ResponseWriter, r *http.Request)
}

type historyController struct {
	hr repository.HistoryRepository
}

func NewHistoryController(hr repository.HistoryRepository) HistoryController {
	return &historyController{hr: hr}
}

func (hc *historyController) GetHistories(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	histories, err := hc.hr.GetHistories(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	var historiesResponse []dto.HistoryRequest
	for _, v := range histories {
		historiesResponse = append(historiesResponse, dto.HistoryRequest{User: v.UserID, Amount: v.Amount})
	}

	fmt.Println(historiesResponse)
	output, _ := json.MarshalIndent(historiesResponse, "", "\t\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

func (hc *historyController) PostHistory(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	decoder := json.NewDecoder(r.Body)

	var hr dto.HistoryRequest

	err := decoder.Decode(&hr)
	if err != nil {
		log.Printf("Failed to unmarshal request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)

		return
	}

	err = hc.hr.InsertHistory(ctx, hr.User, hr.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
}
