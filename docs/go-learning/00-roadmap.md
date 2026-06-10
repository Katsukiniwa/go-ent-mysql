# 00. ロードマップ

## このハンズオンの世界観

「商品 (Product) を在庫管理し、ユーザー (User) が購入すると履歴 (History) が残る」小さな EC バックエンドを、20 トピックかけて育てていく。各お題は **必ず本番コードに 1 つ以上の機能** を足す。

最終形のイメージ:

```
[Client]
   ↓ HTTP
[net/http handlers]            ← トピック03/04/06/09/12 でAPI追加
   ↓
[usecase 層]                   ← トピック05/10/15/16/20 で組み立て
   ↓
[repository interface]         ← トピック05/17 で抽象化
   ↓
[ent client → MySQL]           ← トピック08/17 でスキーマ拡張
```

## TS と Go の対応表

頭の中で常に持っておく:

| TypeScript | Go |
| --- | --- |
| `interface User { id: number }` | `type User struct { ID int }` |
| `class Foo { bar() {} }` | `type Foo struct {}` + `func (f *Foo) Bar() error {}` |
| `implements I` | (なし。メソッドが揃えば暗黙実装) |
| `Promise<T>` | `(T, error)` の多値返り |
| `try / catch (e)` | `if err != nil { return nil, fmt.Errorf("...: %w", err) }` |
| `await Promise.all(xs)` | `errgroup.Group` + `g.Go(...)` |
| `null` / `undefined` | `nil` (型ごとに意味が違う) |
| `JSON.stringify` | `encoding/json` の `json.Marshal` |
| `inversify` / DI コンテナ | `wire` (コード生成型 DI) |
| `Jest` | `testing` パッケージ + テーブル駆動 |
| `pino` | `log/slog` |

## 進め方

1. **ブランチを切る**: `git checkout -b learn/topicNN-short-name`
2. **小さくコミット**: `feat(product): 在庫減算をドメインメソッドに切り出す` のような conventional commit を 1 ステップ 1 コミット。
3. **動作確認の癖**:
   - サーバ起動: `docker compose up -d && go run .`
   - 叩く: `curl -i 'http://localhost:8080/products'`
   - テスト: `go test -race -cover ./...`
   - 静的解析: `go vet ./... && gofmt -s -d .`
4. **詰まったら read source**: `pkg/handler/purchase_handler.go` / `pkg/infrastructure/repository/product/` を読むと答えのヒントがある。

## 進度の目安

- フェーズ 1〜2 (01〜09): 各 30〜60 分。Go の基本構文を体で覚える。
- フェーズ 3 (10〜12): 各 30〜45 分。context が貫通する感覚を掴む。
- フェーズ 4 (13〜16): 各 1〜2 時間。並行処理は必ず `-race` で検証。
- フェーズ 5 (17〜20): 各 1〜3 時間。20 は半日かけて総合演習。

詰まったら飛ばしてもいい。ただし 10 (context) と 17 (ent) はスキップ不可。後続が組めなくなる。

## 完走後の到達点

- handler / usecase / repository の 3 層を、ctx を貫通させて自分で実装できる。
- ent を使った N+1 を回避できる。
- goroutine + errgroup + context.WithTimeout で並行 I/O が書ける。
- Wire でコンパイル時 DI ができる。
- テーブル駆動テストでカバレッジ 80%+ を狙える。
