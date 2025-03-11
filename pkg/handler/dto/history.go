package dto

type HistoryResponse struct {
	ID     int `json:"id"`
	User   int `json:"user"`
	Amount int `json:"amount"`
}

type HistoryRequest struct {
	User   int `json:"user"`
	Amount int `json:"amount"`
}

type HistoryListResponse struct {
	HistoryList []HistoryResponse `json:"histories"`
}
