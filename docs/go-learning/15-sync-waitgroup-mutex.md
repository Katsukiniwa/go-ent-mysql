# 15. sync.WaitGroup / sync.Mutex

## このお題で足す機能

`POST /purchases/bulk` を実装。複数商品の一括購入を JSON 配列で受け取り、並行に在庫減算する。in-memory の在庫キャッシュ (`StockCache`) を Mutex で保護。

## ゴール

- `sync.WaitGroup` で goroutine の終了を待つ。
- `sync.Mutex` で in-memory state を保護する。
- `go test -race` で race condition を観察し、Mutex 追加で消す。
- `sync.RWMutex` / `sync.Once` の使い分けを知る。

## TS との違い

TS はシングルスレッドで race の心配が少ない。Go は本物の並行なので、共有 state は明示的に保護が必要。

## ステップ

### Step 1: in-memory な在庫キャッシュ

`pkg/cache/stock_cache.go`:

```go
package cache

import (
    "errors"
    "sync"
)

var ErrOutOfStock = errors.New("out of stock")

type StockCache struct {
    mu    sync.Mutex
    stock map[int]int
}

func NewStockCache() *StockCache {
    return &StockCache{stock: make(map[int]int)}
}

func (c *StockCache) Set(id, n int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.stock[id] = n
}

func (c *StockCache) Decrement(id, n int) (int, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    cur, ok := c.stock[id]
    if !ok {
        return 0, errors.New("not in cache")
    }
    if cur < n {
        return 0, ErrOutOfStock
    }
    c.stock[id] = cur - n
    return c.stock[id], nil
}
```

ポイント:
- `mu` は struct の先頭近くに置く (慣習)。
- `Lock`/`Unlock` のセットは必ず `defer` で。
- Mutex を含む struct は **値コピー禁止** (ポインタレシーバ統一)。

### Step 2: WaitGroup で並行実行

`pkg/usecase/bulk_purchase_usecase.go`:

```go
type BulkItem struct {
    ProductID entity.ProductID
    Amount    int
}

func (u *BulkPurchaseUsecase) Execute(ctx context.Context, items []BulkItem) []error {
    errs := make([]error, len(items))
    var wg sync.WaitGroup
    wg.Add(len(items))

    for i, it := range items {
        i, it := i, it
        go func() {
            defer wg.Done()
            errs[i] = u.purchase(ctx, it.ProductID, it.Amount)
        }()
    }
    wg.Wait()
    return errs
}
```

各 goroutine が **別々のスロット** (`errs[i]`) に書き込むので Mutex 不要。

### Step 3: わざと race を作って観察

```go
func TestRace_NoMutex(t *testing.T) {
    counter := 0
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() { defer wg.Done(); counter++ }()
    }
    wg.Wait()
    t.Logf("counter=%d (1000 にならないことが多い)", counter)
}
```

```bash
go test -race ./pkg/cache/...
# DATA RACE 報告
```

Mutex で修正:

```go
var mu sync.Mutex
go func() {
    defer wg.Done()
    mu.Lock(); counter++; mu.Unlock()
}()
```

**race detector は CI 必須**。

### Step 4: RWMutex で読み多書き少を最適化

```go
type Cache struct {
    mu sync.RWMutex
    m  map[int]int
}

func (c *Cache) Get(k int) int {
    c.mu.RLock(); defer c.mu.RUnlock()
    return c.m[k]
}

func (c *Cache) Set(k, v int) {
    c.mu.Lock(); defer c.mu.Unlock()
    c.m[k] = v
}
```

「読み 100 : 書き 1」のような分布で大幅改善。

### Step 5: sync.Once で初期化を 1 回だけ

```go
var (
    once     sync.Once
    instance *SomeSingleton
)

func Get() *SomeSingleton {
    once.Do(func() { instance = newInstance() })
    return instance
}
```

`init()` 関数より使い分けやすい。panic が起きても `once.Do` は呼び出し完了とみなされる。

### Step 6: handler

```go
func (h *BulkPurchaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var items []usecase.BulkItem
    if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
        http.Error(w, "bad request", 400); return
    }
    errs := h.uc.Execute(r.Context(), items)
    _ = json.NewEncoder(w).Encode(map[string]any{"errors": errs})
}
```

## 完了条件

- `curl -X POST 'localhost:8080/purchases/bulk' -d '[{"ProductID":1,"Amount":1},{"ProductID":2,"Amount":3}]'` で並行実行される。
- `go test -race ./...` 緑。
- Step 3 で race を観察 → Mutex 追加 → 解消の差分が見える。

## 詰まりやすいポイント

- **Mutex を含む struct の値コピー**: `func (c StockCache) ...` のような値レシーバ NG。`go vet` が警告。
- **デッドロック**: 同じ goroutine が同じ Mutex を再帰取得すると死ぬ (Go の Mutex は再入不可)。
- **長すぎる critical section**: Lock 中に I/O や DB クエリを呼ばない。スループット激減。
- **WaitGroup の Add は go の前**: goroutine の中で `wg.Add(1)` を呼ぶと「main が Wait に到達したとき Add 0」で待たずに抜けるバグ。

## 関連パターン

- 各 goroutine が別スロットに書く設計で Mutex を不要にする
- RWMutex で読み多書き少を最適化
- `go test -race` を CI 必須に
- 長い critical section を避ける
