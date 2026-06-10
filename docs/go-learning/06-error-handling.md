# 06. エラーハンドリング

## このお題で足す機能

ドメインエラー (`ErrOutOfStock`, `ErrProductNotFound`, `ErrInvalidInput`) を定義し、handler で `errors.Is` / `errors.As` で判別して HTTP ステータスを出し分ける (404 / 409 / 400 / 500)。

## ゴール

- `if err != nil { return err }` の作法を体に染み込ませる。
- `fmt.Errorf("...: %w", err)` で **wrap** することで原因を保ったまま文脈を足す。
- `errors.Is` (sentinel) と `errors.As` (型) の使い分け。
- panic を業務ロジックで使わない (起動失敗等のみ)。

## TS との違い

```ts
try { await uc.purchase(); }
catch (e) {
  if (e instanceof OutOfStockError) return res.status(409);
  if (e instanceof NotFoundError) return res.status(404);
  return res.status(500);
}
```

```go
err := uc.Purchase(ctx, ...)
switch {
case errors.Is(err, entity.ErrOutOfStock):
    http.Error(w, "out of stock", 409)
case errors.Is(err, entity.ErrProductNotFound):
    http.Error(w, "not found", 404)
case err != nil:
    http.Error(w, "internal", 500)
}
```

## ステップ

### Step 1: ドメインエラーを定義

`pkg/entity/errors.go`:

```go
package entity

import (
    "errors"
    "fmt"
)

var (
    ErrOutOfStock      = errors.New("out of stock")
    ErrProductNotFound = errors.New("product not found")
    ErrUserNotFound    = errors.New("user not found")
)

// フィールド付きは型で
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}
```

**Sentinel error** は値同一性で判別 (`errors.Is`)、**型 error** はフィールド情報を運びたい時 (`errors.As`)。

### Step 2: Repository で wrap

`pkg/infrastructure/repository/product/repository.go`:

```go
func (r *ProductRepository) FindByID(ctx context.Context, id entity.ProductID) (*entity.Product, error) {
    row, err := r.client.Product.Get(ctx, int(id))
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, fmt.Errorf("FindByID(%d): %w", id, entity.ErrProductNotFound)
        }
        return nil, fmt.Errorf("FindByID(%d): %w", id, err)
    }
    return &entity.Product{ID: row.ID, Title: row.Title, Stock: row.Stock}, nil
}
```

`%w` で wrap → 呼び出し側が `errors.Is(err, entity.ErrProductNotFound)` で判定できる。`%s` だと文字列化されてチェーンが切れるので注意。

### Step 3: Usecase のバリデーション

```go
func (u *PurchaseUsecase) Purchase(ctx context.Context, pid entity.ProductID, uid entity.UserID, amount int) error {
    if amount <= 0 {
        return &entity.ValidationError{Field: "amount", Message: "must be positive"}
    }
    p, err := u.products.FindByID(ctx, pid)
    if err != nil {
        return err
    }
    if err := p.DecrementStock(amount); err != nil {
        return err // ErrOutOfStock
    }
    if err := u.products.Save(ctx, p); err != nil {
        return fmt.Errorf("save product %d: %w", pid, err)
    }
    return u.histories.Create(ctx, &entity.History{ProductID: pid, UserID: uid, Amount: amount})
}
```

### Step 4: Handler でステータス分岐

```go
func (h *PurchaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ... parse params ...
    err := h.uc.Purchase(r.Context(), pid, uid, amount)

    var ve *entity.ValidationError
    switch {
    case err == nil:
        w.WriteHeader(http.StatusNoContent)
    case errors.As(err, &ve):
        http.Error(w, ve.Error(), http.StatusBadRequest)        // 400
    case errors.Is(err, entity.ErrProductNotFound):
        http.Error(w, "product not found", http.StatusNotFound) // 404
    case errors.Is(err, entity.ErrOutOfStock):
        http.Error(w, "out of stock", http.StatusConflict)      // 409
    default:
        log.Printf("internal error: %v", err) // サーバ側にだけ詳細
        http.Error(w, "internal", http.StatusInternalServerError)
    }
}
```

`errors.As` は型にマッチし `*ve` にバインドして詳細フィールドにアクセスできる。
`errors.Is` は値の同一性 (チェーン上に同じ sentinel error があるか) を見る。

### Step 5: テストでエラー判定を検証

```go
func TestPurchase_OutOfStock_ReturnsSentinel(t *testing.T) {
    pr := &stubProductRepo{products: map[entity.ProductID]*entity.Product{
        1: {Stock: 1},
    }}
    uc := NewPurchaseUsecase(pr, &stubHistoryRepo{})
    err := uc.Purchase(context.Background(), 1, 100, 5)
    if !errors.Is(err, entity.ErrOutOfStock) {
        t.Fatalf("ErrOutOfStock chain を期待: %v", err)
    }
}
```

## 完了条件

- `curl -i 'localhost:8080/products/9999/purchase?amount=1'` → **404**
- `curl -i 'localhost:8080/products/1/purchase?amount=99999'` → **409**
- `curl -i 'localhost:8080/products/1/purchase?amount=-1'` → **400**
- サーバログで内部エラー詳細を観察 (クライアントには返さない)。

## 詰まりやすいポイント

- **`%s` と `%w` の違い**: `%s` は wrap しない (`errors.Is` で見えなくなる)。新規開発は基本 `%w`。
- **エラー silent 捨て**: `_ = doSomething()` は最終手段。基本ログするか上に返す。
- **ログとエラー返却の二重化**: 各層でログを出すと運用時にノイズ。原則 **最上位 (handler) でログ**、下層は wrap して返すだけ。
- **panic はバグの合図**: 業務エラーで panic を投げない。

## 関連パターン

- Error wrapping with `%w`
- Sentinel errors / custom error types
- `errors.Is` / `errors.As`
- Return early on errors (happy path をインデント浅く保つ)
