package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/katsukiniwa/go-ent-mysql/product/ent"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/handler"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/infrastructure/repository"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/infrastructure/router"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	hello := []byte("pong")

	_, err := w.Write(hello)
	if err != nil {
		log.Fatal(err)
	}
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	hello := []byte("Hello World!")

	_, err := w.Write(hello)
	if err != nil {
		log.Fatal(err)
	}
}

func timeHandler(w http.ResponseWriter, _ *http.Request) {
	ct := time.Now().Format("2006-01-02 15:04:05")

	_, err := w.Write([]byte(ct))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// @link https://qiita.com/masa_ito/items/66571c993e53eee37ff3
	// ポインタの練習場所↓
	i := 1
	i2 := i
	p := &i
	fmt.Println(i)  // => 1
	fmt.Println(i2) // => 1
	fmt.Println(*p) // => 1

	i2 = 99

	fmt.Println(i)  // => 1
	fmt.Println(i2) // => 99
	fmt.Println(*p) // => 1
	*p = 99

	fmt.Println(i)  // => 99
	fmt.Println(i2) // => 99
	fmt.Println(*p) // => 99
	// ポインタの練習場所↑

	// @link https://zenn.dev/masamiki/articles/83a8db3f132fcb1c48f0
	entOptions := []ent.Option{}
	entOptions = append(entOptions, ent.Debug())
	mc := mysql.Config{
		User:                 "root",
		Passwd:               "password",
		Net:                  "tcp",
		Addr:                 "db" + ":" + "3306",
		DBName:               "product",
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	client, err := ent.Open("mysql", mc.FormatDSN(), entOptions...)

	tr := repository.NewProductRepository(client)

	tc := handler.NewGetProductsHandler(tr)

	pc := handler.NewPurchaseHandler(tr)

	hr := repository.NewHistoryRepository(client)

	hc := handler.NewHistoryController(hr)

	ro := router.NewRouter(tc, pc, hc)

	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	}

	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Printf("failed creating schema resources: %v", err)

		return
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Println("Error closing client:", err)
		}
	}()

	var to = 10 * time.Second

	server := http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: to,
	}

	http.HandleFunc("/products", ro.HandleProductsRequest)
	http.HandleFunc("/histories", ro.HandleHistoriesRequest)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/time", timeHandler)
	http.HandleFunc("/", rootHandler)

	log.Println("Server starting on :8080")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
