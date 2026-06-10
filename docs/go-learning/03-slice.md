# 03. slice

## このお題で足す機能

`GET /products?inStock=true&sort=title` を実装する。在庫がある商品だけを抽出し、タイトル昇順で返す。

## ゴール

- `len` / `cap` / 内部配列のメンタルモデルを掴む。
- `append` が**元のスライスを書き換えるケース**と**新しい配列に切り替わるケース**を予測できる。
- 配列 (`[N]T`) と slice (`[]T`) の差を言える。
- `sort.Slice` でソート関数を書ける。

## TS との違い

```ts
// TS: 配列は実質スライス。コピー渡し時もリファレンス共有。
const a = [1, 2, 3, 4, 5];
const b = a.slice(1, 3); // 新規配列 [2, 3]
b[0] = 99;               // a は変わらない
```

```go
// Go: slice は配列へのビュー。同じ内部配列を共有することがある。
a := []int{1, 2, 3, 4, 5}
b := a[1:3]
b[0] = 99
// a は変わる! → [1, 99, 3, 4, 5]
```

## ステップ

### Step 1: フィルタ関数を Entity に追加

`pkg/entity/product_filter.go`:

```go
package entity

import "sort"

type ProductFilter struct {
    InStockOnly bool
    SortByTitle bool
}

func ApplyProductFilter(products []*Product, f ProductFilter) []*Product {
    out := make([]*Product, 0, len(products)) // ← cap を予約 (アロケ削減)
    for _, p := range products {
        if f.InStockOnly && p.Stock <= 0 {
            continue
        }
        out = append(out, p)
    }
    if f.SortByTitle {
        sort.Slice(out, func(i, j int) bool { return out[i].Title < out[j].Title })
    }
    return out
}
```

`make([]*Product, 0, len(products))` で **cap を予約**しているのがポイント。append の都度の reallocate を避ける (golang-patterns: preallocate slices)。

### Step 2: handler でクエリパラメータを読む

`pkg/handler/get_products_handler.go` を編集:

```go
func (h *GetProductsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    products, err := h.repo.FindAll(ctx)
    if err != nil { http.Error(w, err.Error(), 500); return }

    filter := entity.ProductFilter{
        InStockOnly: r.URL.Query().Get("inStock") == "true",
        SortByTitle: r.URL.Query().Get("sort") == "title",
    }
    filtered := entity.ApplyProductFilter(toEntities(products), filter)

    _ = json.NewEncoder(w).Encode(filtered)
}
```

### Step 3: slice 共有のハマりどころを観察

`pkg/entity/product_filter_test.go` に「共有確認テスト」を:

```go
func TestSlice_SharesUnderlyingArray(t *testing.T) {
    base := []int{1, 2, 3, 4, 5}
    view := base[1:3]
    view[0] = 99
    if base[1] != 99 {
        t.Fatal("base[1] も書き換わっているはず")
    }
}

func TestAppend_BehaviorDependsOnCap(t *testing.T) {
    s := make([]int, 0, 4)
    s1 := append(s, 1)
    s2 := append(s, 2) // cap が足りるので s1 を上書きしてしまう
    if s1[0] != 2 {
        t.Logf("注意: append は cap が足りるとき内部配列を共有する。s1[0]=%d", s1[0])
    }
}
```

### Step 4: バルク操作のヘルパ

ent から取った `[]*ent.Product` を `[]*entity.Product` に変換する `toEntities` 関数を `pkg/handler/converter.go` に作る。preallocate を必ず使う。

```go
func toEntities(src []*ent.Product) []*entity.Product {
    out := make([]*entity.Product, 0, len(src))
    for _, p := range src {
        out = append(out, &entity.Product{
            ID: p.ID, Title: p.Title, Stock: p.Stock,
        })
    }
    return out
}
```

## 完了条件

- `curl 'localhost:8080/products?inStock=true&sort=title'` で在庫>0 のみがタイトル順で返る。
- `go test ./pkg/entity/... -v` 緑、share/append テストの意味を理解。

## 詰まりやすいポイント

- **slice の引数渡し**: 関数に slice を渡すと **slice ヘッダ自体はコピー**されるが、内部配列は共有。`append` で長さを変えても呼び出し元には伝わらない。
- **for range のコピー**: `for _, p := range products` の `p` は **要素のコピー**。書き換えたければ `for i := range products { products[i].X = ... }`。
- **nil slice と空 slice**: `var s []int` (nil) と `[]int{}` (空) は `len == 0` だが `== nil` の結果が違う。JSON エンコード結果も `null` vs `[]` で異なる。

## 関連パターン

- Preallocate slices when size is known
- Return concrete types (ApplyProductFilter は `[]*Product` を返す)
