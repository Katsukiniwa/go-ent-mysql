# 14. channel

## このお題で足す機能

在庫補充イベントを購読する `StockNotifier` を実装。`POST /products/{id}/restock` で在庫追加 → 内部 channel 経由でリスナー (ロガー / Slack 通知モック) に配信する。Worker Pool で並行配信。

## ゴール

- `chan T` / `chan<- T` / `<-chan T` を使い分ける。
- バッファ無し / バッファ付き / close / range の 4 つを書ける。
- **goroutine リーク** を `select + ctx.Done` で防ぐ。
- Worker Pool パターンを実装できる。

## TS との違い

TS には標準で channel はない。EventEmitter / Observable が近いが、Go の channel は型付き・サイズ指定可・close 概念ありで低レベル。

## ステップ

### Step 1: イベント型と Notifier を定義

`pkg/notifier/stock.go`:

```go
package notifier

type StockRestocked struct {
    ProductID int
    Added     int
    NewStock  int
}

type StockNotifier struct {
    ch chan StockRestocked
}

func NewStockNotifier(buf int) *StockNotifier {
    return &StockNotifier{ch: make(chan StockRestocked, buf)}
}

func (n *StockNotifier) Publish(ev StockRestocked) {
    n.ch <- ev // バッファ満杯ならブロック
}

func (n *StockNotifier) Subscribe() <-chan StockRestocked {
    return n.ch // 受信専用として渡す
}

func (n *StockNotifier) Close() { close(n.ch) }
```

`<-chan T` で公開すると **外部からは受信しかできない**。型レベルで方向を絞れるのが Go channel の強み。

### Step 2: Worker Pool で配信

```go
func StartLoggers(ctx context.Context, n *StockNotifier, workers int) {
    sub := n.Subscribe()
    for i := 0; i < workers; i++ {
        i := i
        go func() {
            for {
                select {
                case ev, ok := <-sub:
                    if !ok {
                        return // channel が close された
                    }
                    log.Printf("[worker-%d] product %d +%d => %d", i, ev.ProductID, ev.Added, ev.NewStock)
                case <-ctx.Done():
                    return
                }
            }
        }()
    }
}
```

ポイント:
- `for { select { ... } }` がイベントループの定石。
- `ev, ok := <-sub` の `ok` が false なら channel が close されたサイン。
- `ctx.Done()` を一緒に監視することで goroutine リークを防ぐ。

### Step 3: handler から Publish

```go
type RestockHandler struct {
    uc       *usecase.RestockUsecase
    notifier *notifier.StockNotifier
}

func (h *RestockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    pid, _ := strconv.Atoi(r.PathValue("id"))
    add, _ := strconv.Atoi(r.URL.Query().Get("amount"))

    newStock, err := h.uc.Restock(ctx, entity.ProductID(pid), add)
    if err != nil { http.Error(w, "error", 500); return }

    h.notifier.Publish(notifier.StockRestocked{
        ProductID: pid, Added: add, NewStock: newStock,
    })
    w.WriteHeader(http.StatusNoContent)
}
```

### Step 4: 送信のデッドロックを観察

```go
ch := make(chan int) // バッファなし
ch <- 1              // 受信者がいないので永遠にブロック → デッドロック
```

`fatal error: all goroutines are asleep - deadlock!` を実体験。バッファ付きにするか送信前に受信側 goroutine を起動する。

### Step 5: close と range の組み合わせ

```go
func produce(ch chan<- int) {
    for i := 1; i <= 5; i++ { ch <- i }
    close(ch) // これがないと consumer が永遠に待つ
}

func consume(ch <-chan int) {
    for v := range ch {
        fmt.Println(v)
    }
}
```

**close するのは送信側のみ**。受信側が close すると panic。

### Step 6: main.go で組み立て

```go
n := notifier.NewStockNotifier(64)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
notifier.StartLoggers(ctx, n, 3)

restockHandler := handler.NewRestockHandler(restockUC, n)
mux.Handle("POST /products/{id}/restock", restockHandler)
```

## 完了条件

- `curl -X POST 'localhost:8080/products/1/restock?amount=10'` でログに `[worker-N] product 1 +10 => ...` が出る。
- サーバ停止 (Ctrl+C) で worker goroutine が `ctx.Done()` 経由で終了する。
- `go test -race ./...` 緑。
- バッファサイズを 0 に変えて、受信者がいないとデッドロックすることを観察。

## 詰まりやすいポイント

- **nil channel への send/recv は永久ブロック**: `var ch chan int; ch <- 1` でフリーズ。
- **close 済み channel への send は panic**: 一度 close したら二度と書かない。close の責任は送信側だけが持つ設計に。
- **多対多 close**: 複数 goroutine が同じ channel を送信する場合、close タイミングが難しい。`sync.Once` か別の終了 signal channel を使う。
- **buffered channel の容量決め**: 小さすぎるとデッドロックリスク、大きすぎるとメモリ無駄。経験則は「ピークの 2 倍」。

## 関連パターン

- 送信専用 / 受信専用 channel で API を絞る
- イベントループ `for { select { case <-ch: ... case <-ctx.Done(): return } }`
- close は送信側の責任
- Worker Pool で並行配信
