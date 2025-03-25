package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

const (
	baseURL     = "http://localhost:8080" // テスト対象サーバー
	amountLimit = 100000                  // ユーザーあたりの最大出金金額
)

type History struct {
	Id     int `json:"id"`
	User   int `json:"user"`
	Amount int `json:"amount"`
}

// 出金リクエストを同時に行なった際に最大出金金額を超えていないかを確かめるためのテスト.
func TestCreateHistory(t *testing.T) {
	conn, err := sql.Open("mysql", "root:password@(localhost:3306)/product")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := conn.Exec("TRUNCATE histories"); err != nil {
		t.Fatal(err)
	}

	// 並列で取引登録リクエストをPOSTする
	var wg sync.WaitGroup

	for i := range 4 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := range 6 {
				uID := (i+j)%2 + 1 // テスト対象のユーザーID。1か2のいずれか。

				req, err := request(uID)
				if err != nil {
					t.Error(err)

					return
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					t.Error(err)

					return
				}

				// ステータスコードの確認
				require.Truef(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest, "unexpected status code %d", resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Error(err)

					return
				}

				t.Log(string(body))

				if err := resp.Body.Close(); err != nil {
					t.Error(err)

					return
				}
			}
		}()
	}

	wg.Wait()

	// 1ユーザあたりの最大出金金額を超えていないか確認
	for _, uID := range []int{1, 2} {
		var amount int
		if err := conn.QueryRow("SELECT IFNULL(SUM(amount), 0) FROM histories WHERE user_id=?", uID).
			Scan(&amount); err != nil {
			t.Fatal(err)
		}

		t.Log(uID, amount)
		require.LessOrEqualf(t, amount, amountLimit, "user:%d amount %d over the amountLimit %d", uID, amount, amountLimit)
	}

	if _, err := conn.Exec("TRUNCATE histories"); err != nil {
		t.Fatal(err)
	}
}

func request(uID int) (*http.Request, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 128))
	if err := json.NewEncoder(buffer).Encode(History{
		User:   uID,
		Amount: 60000,
	}); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/histories",
		buffer,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
