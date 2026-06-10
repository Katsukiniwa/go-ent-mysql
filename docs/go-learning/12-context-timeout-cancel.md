# 12. context.WithTimeout / WithCancel

## このお題で足す機能

すべての repository の DB アクセスに **3 秒タイムアウト** を入れ、超過時は handler が **HTTP 504** を返す。クライアント切断時にも DB クエリが中断されることを確認する。

## ゴール

- `context.WithTimeout` / `WithCancel` で **派生 context** を作れる。
- `defer cancel()` が必須な理由を理解 (リソースリーク防止)。
- `errors.Is(err, context.DeadlineExceeded)` でタイムアウト判別。

## TS との違い

```ts
const ac = new AbortController();
setTimeout(() => ac.abort(), 3000);
const r = await fetch(url, { signal: ac.signal });
```

```go
ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
defer cancel()
rows, err := db.QueryContext(ctx, "SELECT ...")
```

## ステップ

### Step 1: Repository に timeout を入れる

```go
func (r *ProductRepository) FindByID(ctx context.Context, id entity.ProductID) (*entity.Product, error) {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel() // 必須: ctx 終了で goroutine リソースを解放

    row, err := r.client.Product.Get(ctx, int(id))
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("FindByID(%d) timed out: %w", id, err)
        }
        return nil, fmt.Errorf("FindByID(%d): %w", id, err)
    }
    return toEntity(row), nil
}
```

タイムアウト値は **環境変数か config で外出し** が本来 (この演習では即値で OK)。

### Step 2: Handler でタイムアウトを 504 にする

```go
products, err := h.uc.FindAll(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        http.Error(w, "gateway timeout", http.StatusGatewayTimeout) // 504
        return
    }
    http.Error(w, "internal", http.StatusInternalServerError)
    return
}
```

### Step 3: わざと遅い handler を作る

`pkg/handler/slow_handler.go`:

```go
func (h *Handler) slow(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    select {
    case <-time.After(10 * time.Second):
        w.Write([]byte("done"))
    case <-ctx.Done():
        // クライアント切断 or タイムアウト
        h.log.WarnContext(ctx, "client canceled", "err", ctx.Err())
        return
    }
}
```

`select` で `ctx.Done()` を監視するのが **長時間処理の必修パターン**。

### Step 4: クライアント切断を観察

```bash
# ターミナル1
curl --max-time 1 'localhost:8080/slow'
# ターミナル2 (サーバログ)
# WARN client canceled err="context canceled"
```

- `context.Canceled` → 手動キャンセル (クライアント切断含む)
- `context.DeadlineExceeded` → タイムアウト

`errors.Is` で見分ける。

### Step 5: cancel を呼ばないリークを観察

`defer cancel()` を **わざと書かず**にテスト:

```bash
go vet ./...
# vet が warn: "lostcancel: the cancel function returned by ... should be called"
```

`go vet` が静的解析で見つけてくれる。CI で `go vet ./...` を必ず流す。

### Step 6: WithCancel で手動制御

```go
ctx, cancel := context.WithCancel(ctx)
go func() {
    if shouldStop() {
        cancel() // 任意のタイミングで停止
    }
}()
```

WithTimeout は時間トリガ、WithCancel はイベントトリガ。

## 完了条件

- 重い `slow` handler を `curl --max-time 1` で叩くとサーバが即終了 (ctx canceled ログ)。
- DB を意図的に止めた状態で `curl 'localhost:8080/products'` を叩くと 504 が返る (3 秒で諦める)。
- `go vet ./...` が "lostcancel" warning ゼロ。

## 詰まりやすいポイント

- **`defer cancel()` 忘れ**: timer goroutine が残ってリーク。`go vet` で発見可能だが見落としがち。
- **タイムアウトのネスト**: 外側 5s, 内側 3s なら **内側が先に切れる**。短い方が勝つ。
- **HTTP server の write timeout**: server 側にも `ReadHeaderTimeout` / `WriteTimeout` を設定 (main.go の既存設定参照)。アプリ層 ctx と二段構え。
- **DB ドライバの timeout 効かない**: 古い `database/sql` ドライバには ctx 対応が無いものがある。ent + mysql は対応済み。

## 関連パターン

- `defer cancel()` は必須
- `select { case <-ctx.Done(): ... case <-time.After(...): ... }`
- 静的解析 (`go vet`) を CI に組み込む
