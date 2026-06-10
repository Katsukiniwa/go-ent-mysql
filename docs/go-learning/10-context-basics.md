# 10. context 基礎

## このお題で足す機能

すべての handler → usecase → repository のメソッドに `ctx context.Context` を**第一引数**として貫通させる。クライアント切断 (`r.Context().Done()`) で DB クエリも中断する構造にする。

## ゴール

- context が「リクエスト固有のキャンセル / タイムアウト / 値の伝搬」を担う仕組みを理解。
- 「ctx は第一引数、construct しない、struct に持たせない」3 原則を体得。
- `context.Background()` と `context.TODO()` の使い分け。

## TS との違い

TS には標準の "context" 概念がなく、最も近いのは `AbortController`:

```ts
const ac = new AbortController();
fetch(url, { signal: ac.signal });
setTimeout(() => ac.abort(), 3000);
```

Go の `ctx` はこれの拡張版 + 値伝搬 + チェーン可能。

## ステップ

### Step 1: 全層に ctx を貫通させる

`pkg/infrastructure/repository/product/repository.go`:

```go
// Before
func (r *ProductRepository) FindByID(id int) (*entity.Product, error)

// After
func (r *ProductRepository) FindByID(ctx context.Context, id entity.ProductID) (*entity.Product, error) {
    row, err := r.client.Product.Get(ctx, int(id)) // ent も ctx を受ける
    ...
}
```

usecase / handler も同様。`ctx` は **必ず第一引数**。

### Step 2: handler はリクエストの context を使う

`pkg/handler/purchase_handler.go`:

```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // クライアント切断で Done() が閉じる
    if err := h.uc.Purchase(ctx, pid, uid, amount); err != nil { ... }
}
```

`r.Context()` は net/http が用意してくれる。クライアント切断で自動キャンセル。

### Step 3: ctx をどこに使うか

ent / `database/sql` / `http.Client` / その他 I/O は **必ず ctx を受ける API** がある。受けないバージョンを使うと**キャンセルが効かない**。

```go
// Bad
rows, err := db.Query("SELECT ...")
// Good
rows, err := db.QueryContext(ctx, "SELECT ...")
```

### Step 4: Background と TODO の使い分け

```go
ctx := context.Background() // アプリのトップレベル (main, テスト)
ctx := context.TODO()       // 将来 ctx を引き回すべきだが今は無い場所
```

`TODO` を使うと grep でリファクタ漏れを発見しやすい。

### Step 5: ctx を構造体に保持しないルール

```go
// Bad
type BadHandler struct { ctx context.Context } // アンチパターン

// Good
type Handler struct { uc *usecase.PurchaseUsecase }
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    h.uc.Do(ctx, ...)
}
```

理由: context は **リクエストスコープ**で、struct のライフタイム (=アプリ全体) と寿命が違う。

### Step 6: テストで Background を使う

```go
func TestPurchase(t *testing.T) {
    err := uc.Purchase(context.Background(), 1, 100, 3)
}
```

タイムアウト挙動の検証は次トピック。

## 完了条件

- `pkg/handler/`, `pkg/usecase/`, `pkg/infrastructure/repository/` の全公開メソッドが `ctx context.Context` を第一引数に取る。
- `grep -r "context.TODO" pkg/` がゼロ。
- 重い handler (sleep 3 秒等) を curl で `--max-time 1` で叩いたとき、サーバ側に「ctx canceled」ログが出る (次トピックで実装)。

## 詰まりやすいポイント

- **ctx を nil で渡す**: 必ず `Background()` か `TODO()`。nil 渡しは panic の元。
- **ctx を変数に格納して使い回す**: ctx は短命。リクエストごとに新規。
- **ctx 無し API を混在**: `db.Query()` / `tx.Commit()` のような ctx 無し版が混在しがち。全部 `QueryContext` / `CommitContext` に揃える。
- **長時間ループ**: ループ内で長く処理するときは `select { case <-ctx.Done(): return ctx.Err(); default: }` を入れる癖。

## 関連パターン

- ctx は第一引数 (Go の絶対ルール)
- ctx を struct に保持しない
- ctx 無し API を使わない (I/O 系)
