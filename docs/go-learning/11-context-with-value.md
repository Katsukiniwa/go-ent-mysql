# 11. context.WithValue

## このお題で足す機能

ミドルウェアで `X-Request-ID` を生成して context に詰め、handler / usecase / repository のログ出力に必ず request-id を載せる。`log/slog` の構造化ログを導入。

## ゴール

- `context.WithValue` の **適切な使い所** を判断できる (ビジネス引数として濫用しない)。
- カスタム key 型で衝突を防ぐイディオム。
- `log/slog` で context-aware なログを書く。

## TS との違い

```ts
const als = new AsyncLocalStorage();
als.run({ requestId: "..." }, () => handler(req, res));
```

```go
ctx = context.WithValue(ctx, requestIDKey{}, "abc-123")
// 下層: ctx.Value(requestIDKey{}).(string)
```

## ステップ

### Step 1: key 型を定義 (衝突防止)

`pkg/middleware/context_keys.go`:

```go
package middleware

import "context"

type requestIDKey struct{}
type userIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestID(ctx context.Context) string {
    v, _ := ctx.Value(requestIDKey{}).(string)
    return v
}
```

**`string` をそのまま key にしない**。型を切ることで他パッケージとの衝突をコンパイラレベルで防ぐ。

### Step 2: ミドルウェアで request-id を生成

`pkg/middleware/request_id.go`:

```go
package middleware

import (
    "net/http"

    "github.com/google/uuid"
)

func RequestIDMW(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.NewString()
        }
        w.Header().Set("X-Request-ID", id)
        ctx := WithRequestID(r.Context(), id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

`r.WithContext(ctx)` で新しい request を作って下流に渡す。

### Step 3: slog のハンドラを context 対応に

`pkg/middleware/logger.go`:

```go
type ctxHandler struct{ slog.Handler }

func (h *ctxHandler) Handle(ctx context.Context, r slog.Record) error {
    if id := RequestID(ctx); id != "" {
        r.AddAttrs(slog.String("request_id", id))
    }
    return h.Handler.Handle(ctx, r)
}

func NewLogger() *slog.Logger {
    base := slog.NewJSONHandler(os.Stdout, nil)
    return slog.New(&ctxHandler{Handler: base})
}
```

### Step 4: usecase / repository でログ

```go
func (u *PurchaseUsecase) Purchase(ctx context.Context, ...) error {
    u.log.InfoContext(ctx, "purchase started", "pid", pid, "amount", amount)
    ...
}
```

**`InfoContext`** を使うことで ctx の値がログに自動付与される。

### Step 5: main.go で組み立て

```go
mux := http.NewServeMux()
// handlers register...

handler := middleware.RequestIDMW(mux)
server := &http.Server{Handler: handler, Addr: ":8080"}
```

### Step 6: 動作確認

```bash
curl -i -H 'X-Request-ID: abc-123' 'localhost:8080/products'
# サーバログ:
# {"time":"...","level":"INFO","msg":"purchase started","pid":1,"request_id":"abc-123"}
```

## 完了条件

- handler→usecase→repository の全層のログに **同一 request_id** が載る。
- レスポンスヘッダに `X-Request-ID` が返る。
- key 型が `string` ではなく非公開型 (`grep -r 'context.WithValue(ctx, "' pkg/` がゼロ)。

## 詰まりやすいポイント

- **`string` を key に使う**: 他パッケージと衝突する。常に非公開型。
- **WithValue を「引数の代わり」に使う**: ビジネスデータ (商品 ID 等) は **明示的に引数で**渡す。WithValue は request-scoped メタデータ (request-id, auth user, tracing span) 専用。
- **型アサーション失敗**: `ctx.Value(...).(string)` のゼロ値 (`""`) と「存在しない」が区別できない。2 値受け `v, ok := ctx.Value(...).(string)` を使う。

## 関連パターン

- カスタム key 型 (`type fooKey struct{}`)
- `log/slog` の context-aware handler
- WithValue は metadata のみ、ビジネス引数は通常引数
