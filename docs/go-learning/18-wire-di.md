# 18. Wire (DI)

## このお題で足す機能

肥大化した `main.go` の手動配線を `github.com/google/wire` に置き換える。`wire.go` (ProviderSet) を書いて `wire_gen.go` を自動生成し、`main.go` を 10 行程度に圧縮する。

## ゴール

- **コンパイル時 DI** のメリットを実感 (実行時リフレクション無し、起動が高速)。
- ProviderSet / Inject / Bind の 3 概念を使い分ける。
- インタフェースと具象の bind 方法。

## TS との違い

```ts
// tsyringe / inversify は decorator + 実行時リフレクション
@injectable() class FooService {}
container.register(FooService).resolve(FooService);
```

```go
// Wire: コード生成。実行時オーバーヘッドゼロ
// コンパイル時に依存関係エラーが出る
```

## ステップ

### Step 1: Wire を入れる

```bash
go install github.com/google/wire/cmd/wire@latest
go get github.com/google/wire
```

### Step 2: ProviderSet を書く

`cmd/server/wire.go`:

```go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
)

func InitializeServer() (*http.Server, error) {
    wire.Build(
        // インフラ
        NewEntClient,
        NewLogger,

        // Repositories
        repository.NewProductRepository,
        repository.NewHistoryRepository,
        repository.NewCartRepository,

        // Usecases
        usecase.NewPurchaseUsecase,
        usecase.NewDashboardUsecase,
        usecase.NewCartUsecase,

        // Handlers
        handler.NewPurchaseHandler,
        handler.NewDashboardHandler,
        handler.NewCartHandler,

        // Router + Server
        NewRouter,
        NewHTTPServer,

        // interface → 具象の bind
        wire.Bind(new(usecase.ProductRepository), new(*repository.ProductRepository)),
        wire.Bind(new(usecase.HistoryRepository), new(*repository.HistoryRepository)),
    )
    return nil, nil
}
```

ポイント:
- `//go:build wireinject` で **このファイルは wire 専用 (本ビルドからは除外)**。
- `wire.Bind` でインタフェースに具象実装を結びつける。
- 戻り値は `nil, nil` (wire が wire_gen.go に上書き)。

### Step 3: Provider 関数を定義

`cmd/server/providers.go`:

```go
func NewEntClient() (*ent.Client, error) {
    mc := mysql.Config{ /* ... */ }
    return ent.Open("mysql", mc.FormatDSN())
}

func NewLogger() *slog.Logger {
    return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func NewRouter(
    p *handler.PurchaseHandler,
    d *handler.DashboardHandler,
    c *handler.CartHandler,
) *http.ServeMux {
    mux := http.NewServeMux()
    mux.Handle("POST /products/{id}/purchase", p)
    mux.Handle("GET /dashboard", d)
    mux.Handle("GET /users/{id}/cart", c)
    return mux
}

func NewHTTPServer(mux *http.ServeMux) *http.Server {
    return &http.Server{
        Addr: ":8080", Handler: mux,
        ReadHeaderTimeout: 10 * time.Second,
    }
}
```

### Step 4: wire コマンドで生成

```bash
cd cmd/server
wire
```

`wire_gen.go` が出力される。中身は手動配線が展開された普通の Go コード。

### Step 5: main.go を圧縮

`cmd/server/main.go`:

```go
package main

import "log"

func main() {
    server, err := InitializeServer()
    if err != nil { log.Fatal(err) }
    log.Println("listening on", server.Addr)
    log.Fatal(server.ListenAndServe())
}
```

**10 行**まで圧縮された。依存配線は `wire_gen.go` (生成物) に集約。

### Step 6: 依存追加時のワークフロー

新しい usecase を作ったら:

1. Provider 関数 (コンストラクタ) を書く。
2. ProviderSet (`wire.go`) に追加。
3. `wire` コマンドで再生成。
4. 不足があるとコンパイルエラー → ProviderSet に足す。

依存関係漏れ・循環を**コンパイル時に発見**できるのが Wire の強み。

## 完了条件

- `cmd/server/wire_gen.go` が生成され、`go build ./cmd/server` 緑。
- 既存の `main.go` の手動配線が消えて 10 行程度。
- 新しい handler を 1 つ追加 → wire 再生成 で動作することを確認。
- `wire_gen.go` をコミットする (生成物だがビルド再現性のため commit が慣習)。

## 詰まりやすいポイント

- **wire コマンド未インストール**: `go install` で PATH に入れる。CI でも実行。
- **bind 忘れ**: インタフェースを受け取る Provider があるのに `wire.Bind` が無いと「no provider found」エラー。
- **重複 Provider**: 同じ型を返す Provider が複数あるとどっちを使うか曖昧でエラー。
- **build tag**: `//go:build wireinject` を忘れると wire.go が本ビルドに入ってしまい二重定義エラー。

## 関連パターン

- コンパイル時 DI でランタイムオーバーヘッドゼロ
- `wire.Bind` で interface → struct
- 生成コードは commit する慣習
