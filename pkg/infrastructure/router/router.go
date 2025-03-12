package router

import (
	"net/http"

	"github.com/katsukiniwa/go-ent-mysql/product/pkg/handler"
)

type Router interface {
	HandleProductsRequest(w http.ResponseWriter, r *http.Request)
	HandleHistoriesRequest(w http.ResponseWriter, r *http.Request)
}

type router struct {
	pc handler.GetProductsHandler
	pp handler.PurchaseHandler
	hc handler.HistoryController
}

func NewRouter(pc handler.GetProductsHandler, pp handler.PurchaseHandler, hc handler.HistoryController) Router {
	return &router{pc: pc, hc: hc, pp: pp}
}

func (ro *router) HandleProductsRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ro.pc.GetProducts(w, r)
	case "POST":
		ro.pp.Purchase(w, r)
	// case "PUT":
	// ro.pc.UpdateProduct(w, r)
	// case "DELETE":
	// ro.pc.DeleteProduct(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ro *router) HandleHistoriesRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ro.hc.GetHistories(w, r)
	case "POST":
		ro.hc.PostHistory(w, r)
	// case "PUT":
	// ro.hc.UpdateProduct(w, r)
	// case "DELETE":
	// ro.hc.DeleteProduct(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
