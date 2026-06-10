# 02. 型を深掘り

## このお題で足す機能

`ProductID` / `UserID` の **専用型** を導入し、`HistoryHandler` / `PurchaseHandler` のシグネチャを差し替える。引数取り違え (User ID と Product ID の混同) をコンパイル時に防ぐ。

## ゴール

- `type X = int` (エイリアス) と `type X int` (型定義) の違いを言える。
- ゼロ値を意図して使い分ける。
- 型アサーション (`v.(T)`) と型スイッチ (`switch v := x.(type)`) を書ける。

## TS との違い

```ts
type ProductID = number;
type UserID    = number;
const swap = (p: ProductID, u: UserID) => {};
swap(123 as UserID, 456 as ProductID); // ❌ TS でも防げない (構造的型付け)
```

```go
type ProductID int
type UserID    int
func Swap(p ProductID, u UserID) {}
Swap(UserID(123), ProductID(456)) // ✅ Go はコンパイルエラー
```

## ステップ

### Step 1: ID 型を定義する

`pkg/entity/ids.go` を新規作成:

```go
package entity

type (
    ProductID int
    UserID    int
    HistoryID int
)

// 文字列クエリパラメータからパースするヘルパ
func ParseProductID(s string) (ProductID, error) {
    n, err := strconv.Atoi(s)
    if err != nil { return 0, fmt.Errorf("ProductID parse %q: %w", s, err) }
    if n <= 0 { return 0, fmt.Errorf("ProductID must be positive, got %d", n) }
    return ProductID(n), nil
}
```

`UserID` / `HistoryID` も同様に。

### Step 2: handler のシグネチャを置き換える

`pkg/handler/purchase_handler.go` で、商品 ID を扱っている箇所 (おそらく `strconv.Atoi` の戻り値 `int`) を `entity.ProductID` に変える。

```go
// before
productID, err := strconv.Atoi(r.URL.Query().Get("id"))
// after
productID, err := entity.ParseProductID(r.URL.Query().Get("id"))
```

repository も `func GetByID(ctx context.Context, id entity.ProductID) (*entity.Product, error)` のように受け取り型を変える。

### Step 3: わざと取り違えてコンパイルエラーを観察する

```go
var uid entity.UserID = 1
prod, _ := repo.GetByID(ctx, uid) // ❌ cannot use uid (type UserID) as type ProductID
```

このエラーが出ることを確認したら、**コメントで「Go は型定義で混同を防げる」と一言書いてコミット**する。

### Step 4: ゼロ値テーブル

`pkg/entity/zero_value_test.go` を新規:

```go
func TestZeroValues(t *testing.T) {
    var pid ProductID
    if pid != 0 { t.Fatal("ProductID のゼロ値は 0") }

    var p *Product
    if p != nil { t.Fatal("*Product のゼロ値は nil") }

    var s []ProductID
    if s != nil || len(s) != 0 { t.Fatal("nil slice は len==0 かつ ==nil") }

    var m map[ProductID]int
    if m != nil { t.Fatal("nil map") }
    // m[1] = 1 // ← これは panic する。試して、初期化して直す。
}
```

### Step 5: 型スイッチで any を分岐

`pkg/handler/debug_handler.go` (新規) に、

```go
func describe(v any) string {
    switch x := v.(type) {
    case entity.ProductID: return fmt.Sprintf("ProductID(%d)", x)
    case entity.UserID:    return fmt.Sprintf("UserID(%d)", x)
    case *entity.Product:  return fmt.Sprintf("*Product{ID:%d, Title:%q}", x.ID, x.Title)
    case []entity.ProductID: return fmt.Sprintf("[]ProductID len=%d", len(x))
    default: return fmt.Sprintf("unknown %T", v)
    }
}
```

を書いて、`GET /debug/types` で適当な値をいくつか出力する handler を追加。

## 完了条件

- `go build ./...` が通る。
- `entity.UserID(1)` を `repo.GetByID` に渡すコードがコンパイルエラーになることをコミットメッセージで言及。
- `go test ./pkg/entity/...` 緑。

## 詰まりやすいポイント

- `int(pid)` で int に**明示変換**できるが、暗黙には混ぜられない。
- `type X = Y` (=つき) はエイリアスで互換、`type X Y` は別型。新規導入は基本後者。
- `nil` map に書き込むと panic、`nil` slice に append は OK (新しい slice が返る)。

## 関連パターン

- Make the zero value useful (ProductID(0) は無効値として扱うルール)
- Errors as values (`ParseProductID` の戻り値で表現)
