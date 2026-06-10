# 09. パッケージと import

## このお題で足す機能

肥大化した `pkg/handler/history_handler.go` を `pkg/handler/history/` ディレクトリ配下のパッケージに分割する。`internal/` ディレクトリで外部公開を制限する作法も導入。

## ゴール

- `package main` と `package <name>` の違い。
- **`internal/` の特別扱い** (同一モジュール外から import 不可) を活用。
- 大文字始まり = 公開 / 小文字始まり = 非公開 のシンプルルール。
- 循環 import を避ける配置。

## TS との違い

```ts
export class HistoryHandler {}
// import { HistoryHandler } from "./history-handler";
```

```go
package history
type Handler struct {}
// import "github.com/.../pkg/handler/history"
// 利用側: history.Handler
```

## ステップ

### Step 1: 既存 history_handler.go を分割

```
pkg/handler/history/
  handler.go      # type Handler とコンストラクタ
  list.go         # GET /histories
  get.go          # GET /histories/{id}
  refundable.go   # GET /histories?refundable=true
```

`pkg/handler/history/handler.go`:

```go
package history

import (
    "net/http"
    "github.com/katsukiniwa/go-ent-mysql/product/pkg/usecase"
)

type Handler struct {
    uc *usecase.HistoryUsecase
}

func NewHandler(uc *usecase.HistoryUsecase) *Handler {
    return &Handler{uc: uc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
    mux.HandleFunc("GET /histories", h.list)
    mux.HandleFunc("GET /histories/{id}", h.get)
}
```

`list.go` / `get.go` には小文字始まり (`list`, `get`) のメソッドだけ書く。同じパッケージ内なので分割しても private に保てる。

### Step 2: internal/ で公開境界を引く

ent client を直接触る低レイヤを `pkg/internal/store/` に動かす:

```
pkg/internal/store/    ← 同モジュール外から import 不可
  ent_helpers.go
```

ルール: **`internal/` 配下は、`internal/` の親ディレクトリと同じプレフィックス配下からしか import できない**。

```go
// pkg/usecase/foo.go から:
import "github.com/.../pkg/internal/store" // OK

// 別モジュールから:
import "github.com/.../pkg/internal/store" // コンパイルエラー
```

### Step 3: package 名と import path の関係

```go
// パス: pkg/handler/history
// 宣言: package history          // 最後の要素と一致 (慣習)
// import: "github.com/.../pkg/handler/history"
// 利用: history.NewHandler(...)
```

**パッケージ名は短く・小文字・アンダースコアなし**。`historyHandler` (×) `history_handler` (×) `history` (○)。

### Step 4: 循環 import を起こしてから直す

わざと `pkg/entity` から `pkg/usecase` を import すると import cycle で詰む。Go は循環を許さないので、依存方向を一方通行にする設計を強制される。

```
handler → usecase → entity
            ↓
        repository → ent
```

実務では `entity` は最下層 (誰も import しない) にする。

### Step 5: main.go を更新

```go
historyHandler := history.NewHandler(historyUC)
historyHandler.Routes(http.DefaultServeMux)
```

main.go から具体的なエンドポイント名 (`/histories`) が消え、Routes に隠蔽される。

## 完了条件

- `go build ./...` 緑。
- `pkg/handler/history_handler.go` が削除され、`pkg/handler/history/*.go` 群に分割されている。
- `internal/store` が同モジュール外から import 不可なことを `go doc` で確認。

## 詰まりやすいポイント

- **package 名 ≠ ディレクトリ名でも動く** が混乱の元。揃える。
- **複数 .go が同 package**: 同ディレクトリ内なら自由に分割できる (TS の "barrel file" 不要)。
- **大文字始まりだけが公開**: フィールド `id int` は JSON marshal でも出ない (タグ無しの場合)。
- **import 整理**: `goimports -w .` で並べ替え & 不要 import 削除。CI に組み込むと楽。

## 関連パターン

- 標準レイアウト `cmd/`, `internal/`, `pkg/`
- パッケージ名は短く一語
- 公開 API は意図的に厳選 (`internal/` を使う)
