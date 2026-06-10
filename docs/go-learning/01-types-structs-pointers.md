# 01. 型・構造体・ポインタ

## このお題で足す機能

`Product` のドメインメソッド `DecrementStock(n int) error` を追加し、`PurchaseHandler` から呼ぶように差し替える。在庫減算ロジックを handler から Entity 層へ寄せる。

## ゴール

- 値レシーバ `func (p Product)` とポインタレシーバ `func (p *Product)` の **挙動の違い** を実際にコードで体験する。
- 関数の引数として「値渡し / ポインタ渡し」を選ぶ判断基準を持つ。
- 「`nil` ポインタ参照で落ちる」失敗を 1 回経験して直す。

## TS との違い (最重要)

```ts
// TS: オブジェクトは常に参照渡し
const p = { stock: 5 };
function decrement(p: { stock: number }, n: number) { p.stock -= n; }
decrement(p, 2);          // p.stock === 3
```

```go
// Go: struct はデフォルト値渡し (= コピーされる)
p := Product{Stock: 5}
func decrementValue(p Product, n int) { p.Stock -= n } // 呼び出し元に反映されない
func decrementPtr(p *Product, n int)  { p.Stock -= n } // 反映される
```

## ステップ

### Step 1: Entity に値オブジェクトを作る

`pkg/entity/product.go` を新規作成 (本リポジトリの `pkg/entity/` には既に他の entity がある。同じスタイルで)。

```go
package entity

import "errors"

var ErrOutOfStock = errors.New("out of stock")

type Product struct {
    ID         int
    Title      string
    Stock      int
    SaleStatus string
}

// 値レシーバ版: コピーに対して変更するため呼び出し元には反映されない
func (p Product) DecrementStockValue(n int) error {
    if p.Stock < n {
        return ErrOutOfStock
    }
    p.Stock -= n
    return nil
}

// ポインタレシーバ版: 元の struct を書き換える
func (p *Product) DecrementStock(n int) error {
    if p.Stock < n {
        return ErrOutOfStock
    }
    p.Stock -= n
    return nil
}
```

### Step 2: 違いを観察するテストを書く

`pkg/entity/product_test.go`:

```go
package entity

import "testing"

func TestDecrementStock_ValueReceiverDoesNotMutate(t *testing.T) {
    p := Product{Stock: 5}
    _ = p.DecrementStockValue(2)
    if p.Stock != 5 {
        t.Fatalf("値レシーバなのに変わってしまった: %d", p.Stock)
    }
}

func TestDecrementStock_PointerReceiverMutates(t *testing.T) {
    p := &Product{Stock: 5}
    if err := p.DecrementStock(2); err != nil {
        t.Fatal(err)
    }
    if p.Stock != 3 {
        t.Fatalf("ポインタなのに変わってない: %d", p.Stock)
    }
}

func TestDecrementStock_OutOfStock(t *testing.T) {
    p := &Product{Stock: 1}
    if err := p.DecrementStock(2); err != ErrOutOfStock {
        t.Fatalf("ErrOutOfStock を期待したが %v", err)
    }
}
```

### Step 3: handler から呼び出す

`pkg/handler/purchase_handler.go` を開いて、現在 handler 内に書かれている在庫減算ロジックを **Entity の `DecrementStock` 呼び出しに差し替える**。ent の `*ent.Product` から自前 `entity.Product` への変換ヘルパを作ってもよい。

### Step 4: nil ポインタを踏みに行く

わざと書いてみる:

```go
var p *entity.Product // nil
p.DecrementStock(1)   // panic: runtime error: invalid memory address
```

`go run .` で落とし、スタックトレースを読む。`if p == nil { return errors.New("product is nil") }` を入れて直す。

## 完了条件

- `go test ./pkg/entity/...` が緑。
- `curl -X POST localhost:8080/products/1/purchase` 相当のリクエストで在庫が減る (実際のエンドポイントは現状 `purchase_handler.go` 参照)。
- handler から在庫減算の `if` 文が消え、`product.DecrementStock(n)` 1 行になっている。

## 詰まりやすいポイント

- **ポインタレシーバと値レシーバの混在**: 1 つの型につき **どちらかに統一**するのが Go の作法。mutating メソッドがあるならポインタレシーバで全部揃える。
- **構造体のコピーコスト**: 小さい struct なら値レシーバでも問題ないが、`sync.Mutex` を含む型は必ずポインタレシーバ (Mutex をコピーすると壊れる)。
- **メソッド呼び出しの自動アドレス取得**: `p.DecrementStock()` は `p` が変数なら `(&p).DecrementStock()` に自動変換される。マップ要素やリテラルではこれが効かないので注意。

## 関連パターン

- Accept interfaces, return structs (今回は struct を扱う側)
- Domain logic を Entity に集約 (handler/usecase は orchestration に専念)
