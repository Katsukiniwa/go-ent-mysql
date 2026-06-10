# 13. goroutine 基礎

## このお題で足す機能

`GET /dashboard` を実装。総売上 / 在庫切れ商品数 / 最近の購入 5 件 を**並行**に取得して、合計レスポンスタイムを最短にする。

## ゴール

- `go f()` で goroutine を立てる感覚を掴む。
- main goroutine が終わると **全 goroutine が即死**することを実体験。
- ループ変数キャプチャ問題 (Go 1.22 以降は緩和) を理解。

## TS との違い

```ts
const [sales, lowStock, recent] = await Promise.all([
  fetchSales(), fetchLowStock(), fetchRecent()
]);
```

```go
ch := make(chan result, 3)
go func() { ch <- fetchSales() }()
go func() { ch <- fetchLowStock() }()
go func() { ch <- fetchRecent() }()
```

Go の goroutine は OS スレッドにスケジュールされる本物の並行。Promise はイベントループのシングルスレッド。

## ステップ

### Step 1: 並行取得 (channel で結果収集)

`pkg/usecase/dashboard_usecase.go`:

```go
type DashboardData struct {
    TotalSales    int
    LowStockCount int
    Recent        []*entity.History
}

func (u *DashboardUsecase) Get(ctx context.Context) (*DashboardData, error) {
    type result struct {
        kind string
        val  any
        err  error
    }
    ch := make(chan result, 3)

    go func() {
        sales, err := u.history.SumAmount(ctx)
        ch <- result{"sales", sales, err}
    }()
    go func() {
        n, err := u.product.CountOutOfStock(ctx)
        ch <- result{"lowStock", n, err}
    }()
    go func() {
        h, err := u.history.RecentN(ctx, 5)
        ch <- result{"recent", h, err}
    }()

    data := &DashboardData{}
    for i := 0; i < 3; i++ {
        r := <-ch
        if r.err != nil {
            return nil, r.err
        }
        switch r.kind {
        case "sales":    data.TotalSales = r.val.(int)
        case "lowStock": data.LowStockCount = r.val.(int)
        case "recent":   data.Recent = r.val.([]*entity.History)
        }
    }
    return data, nil
}
```

ポイント:
- `make(chan result, 3)` は **バッファ付き channel**。3 つの goroutine が結果を送り終わった後に close 不要 (受信側で 3 回読めば終わり)。
- このやり方は雑。次トピックで `errgroup` に置き換える。

### Step 2: handler を追加

```go
func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    data, err := h.uc.Get(r.Context())
    if err != nil {
        http.Error(w, "internal", http.StatusInternalServerError)
        return
    }
    _ = json.NewEncoder(w).Encode(data)
}
```

### Step 3: main goroutine の即死を観察

```go
func main() {
    go func() {
        time.Sleep(2 * time.Second)
        fmt.Println("これは出ない (main が先に終わる)")
    }()
    fmt.Println("main end")
}
```

実行すると `main end` しか出ない。**main goroutine が終わるとプロセスごと終わる**。これが TS の Promise との大きな違い。

### Step 4: ループ変数キャプチャ (Go < 1.22 のハマり)

```go
ids := []int{1, 2, 3}
for _, id := range ids {
    go func() {
        fmt.Println(id) // Go 1.21 以前は全て 3 が出る可能性
    }()
}
```

Go 1.22 以降は各イテレーションで新変数になり修正済み。古典的なシャドウイング `id := id` のテクニックは知識として覚えておく。

### Step 5: -race で競合検出

```bash
go test -race ./pkg/usecase/...
```

並行アクセスがなければ pass。次トピック以降で walking data race を踏むので習慣にする。

## 完了条件

- `curl 'localhost:8080/dashboard'` がレスポンス。
- 3 つのクエリが並行に飛んでいる (DB ログのタイムスタンプが同時)。逐次より速い。
- `go test -race ./...` 緑。

## 詰まりやすいポイント

- **goroutine leak**: goroutine が永遠に channel 待ちでブロックすると leak。次トピックで対策。
- **main の終了で全死**: バックグラウンド処理を待ちたいなら `sync.WaitGroup` か `errgroup`。
- **panic in goroutine**: goroutine 内の panic は **プロセス全体を落とす**。`recover` を入れるか panic させない設計を。

## 関連パターン

- バッファ付き channel で結果集約
- main goroutine の寿命 = プロセスの寿命
- `go test -race` を癖に
