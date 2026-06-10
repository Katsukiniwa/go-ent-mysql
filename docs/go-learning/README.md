# Go ハンズオン (機能追加スタイル)

TypeScript 経験者が、このリポジトリの本番コードに **実際に機能を足しながら** Go を学ぶためのお題集。

20 トピックを既存層 (`ent/schema/`, `pkg/handler/`, `pkg/usecase/`, `pkg/infrastructure/repository/`, `pkg/db/`, `main.go`) に積み上げる形で進める。完走すると「予約作成 + 在庫管理 + 履歴 + 通知 + ユーザー認証」を含む小さな EC バックエンドができあがる。

## 取り組み方

1. お題ごとに `learn/topicNN-...` ブランチを切る (例: `learn/topic03-stock-filter`)。
2. **既存ファイルを編集 or 新規ファイル追加で機能を実装**。`pkg/learn/` のようなサンドボックスは作らない。
3. 動作確認は `go run .` でサーバを立てて `curl` を叩く、もしくは `go test ./...` を走らせる。MySQL は `docker compose up -d` で起動。
4. 各お題の「完了条件」を満たしたら、conventional commit (日本語) で 1 つコミット。

## ロードマップ全体像

各お題で **このリポジトリに足される機能** を 1 行で:

| # | トピック | このリポジトリに足す機能 |
| -- | -- | -- |
| 00 | [ロードマップ](./00-roadmap.md) | (説明) |
| 01 | [型・構造体・ポインタ](./01-types-structs-pointers.md) | `Product.DecrementStock` ドメインメソッドを追加 |
| 02 | [型を深掘り](./02-types-deep-dive.md) | `ProductID` / `UserID` の型定義で handler 引数を型安全化 |
| 03 | [slice](./03-slice.md) | 在庫がある商品だけ返す `GET /products?inStock=true` |
| 04 | [map](./04-map.md) | バルク取得 `GET /products?ids=1,2,3` を map で実装 |
| 05 | [interface](./05-interface.md) | Repository を interface 化して handler から具象を剥がす |
| 06 | [エラーハンドリング](./06-error-handling.md) | `ErrOutOfStock` / `ErrProductNotFound` 独自エラー型と HTTP ステータス分岐 |
| 07 | [メソッドとレシーバー](./07-methods-receivers.md) | `History.IsRefundable()` 等のドメインロジックを Entity に集約 |
| 08 | [defer](./08-defer.md) | Purchase の DB トランザクションを `defer tx.Rollback()` パターンで実装 |
| 09 | [パッケージと import](./09-packages.md) | 肥大化した `history_handler.go` を `pkg/handler/history/` に分割 |
| 10 | [context 基礎](./10-context-basics.md) | handler → usecase → repository に `ctx` を貫通させる |
| 11 | [context.WithValue](./11-context-with-value.md) | request-id をミドルウェアで詰めて全層のログに出す |
| 12 | [context.WithTimeout](./12-context-timeout-cancel.md) | DB アクセスに 3 秒タイムアウト、504 で返す |
| 13 | [goroutine](./13-goroutine.md) | `GET /products/dashboard` で在庫件数 + 売上集計を並行取得 |
| 14 | [channel](./14-channel.md) | 在庫補充イベントを channel で配信する `StockNotifier` |
| 15 | [sync.WaitGroup/Mutex](./15-sync-waitgroup-mutex.md) | バルク購入 API、Mutex で在庫減算を保護 |
| 16 | [context × goroutine](./16-context-goroutine.md) | `errgroup` で並行集計、1 つ失敗で全体キャンセル |
| 17 | [ent (ORM)](./17-ent-orm.md) | `Cart` スキーマと `User ↔ Cart ↔ Product` リレーションを追加 |
| 18 | [Wire (DI)](./18-wire-di.md) | `main.go` の手動配線を Wire で生成、`wire_gen.go` を出力 |
| 19 | [テスト](./19-testing.md) | テーブル駆動テストで `PurchaseUsecase` を網羅 |
| 20 | [総合演習: 予約フロー](./20-integration.md) | `POST /reservations` で在庫減算 + History 記録 + Cart 空にする一連トランザクション |

## 進める上でのお作法

- **conventional commit (日本語) 1 行**: `feat(product): 在庫切れ判定をドメインメソッドに集約`
- **`go vet ./...` と `gofmt -s -w .`** をコミット前に流す (Makefile に追加してもよい)。
- **並行コードは `go test -race`** を必ず通す。
- **既存実装を読む**: `pkg/handler/purchase_handler.go` / `pkg/infrastructure/repository/product/` を写経の素材に。
