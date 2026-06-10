# 04. map

## このお題で足す機能

`GET /products?ids=1,2,3` を実装。複数 ID を一括取得して、`map[ProductID]*Product` で返す。N+1 を避けて 1 クエリで取る。

## ゴール

- `make(map[K]V)` の必要性を理解 (nil map に書くと panic)。
- `v, ok := m[k]` の **2 値受け** で存在チェックする。
- map の **イテレーション順序がランダム**なことを体感。
- ent の `WHERE IN` を使った N+1 回避。

## TS との違い

```ts
// TS: object/Map どちらも未初期化で代入できないわけではない
const m: Record<string, number> = {};
m.bar = 2;
```

```go
// Go: nil map に書くと panic
var m map[string]int
m["bar"] = 1 // panic!

m := make(map[string]int)
m["bar"] = 1 // OK
```

## ステップ

### Step 1: Repository に BulkGet を追加

`pkg/infrastructure/repository/product/repository.go`:

```go
func (r *ProductRepository) FindByIDs(ctx context.Context, ids []entity.ProductID) (map[entity.ProductID]*entity.Product, error) {
    intIDs := make([]int, 0, len(ids))
    for _, id := range ids {
        intIDs = append(intIDs, int(id))
    }

    rows, err := r.client.Product.Query().
        Where(product.IDIn(intIDs...)).
        All(ctx)
    if err != nil {
        return nil, fmt.Errorf("FindByIDs: %w", err)
    }

    out := make(map[entity.ProductID]*entity.Product, len(rows))
    for _, p := range rows {
        out[entity.ProductID(p.ID)] = &entity.Product{
            ID: p.ID, Title: p.Title, Stock: p.Stock,
        }
    }
    return out, nil
}
```

ポイント:
- 返り値を **map** にすると、handler 側の「ID → Product」のルックアップが O(1)。
- `make(map[...]..., len(rows))` で **初期容量を予約**するとリハッシュが減る。

### Step 2: クエリパラメータをパース (重複除去に Set 化)

```go
func parseIDs(s string) ([]entity.ProductID, error) {
    if s == "" {
        return nil, nil
    }
    parts := strings.Split(s, ",")
    ids := make([]entity.ProductID, 0, len(parts))
    seen := make(map[entity.ProductID]struct{}) // Set イディオム
    for _, p := range parts {
        id, err := entity.ParseProductID(strings.TrimSpace(p))
        if err != nil {
            return nil, err
        }
        if _, dup := seen[id]; dup {
            continue
        }
        seen[id] = struct{}{}
        ids = append(ids, id)
    }
    return ids, nil
}
```

`map[K]struct{}` が **Set のイディオム**。値型を `struct{}` にするとメモリ 0 バイト。

### Step 3: handler で分岐

`pkg/handler/get_products_handler.go`:

```go
if idsParam := r.URL.Query().Get("ids"); idsParam != "" {
    ids, err := parseIDs(idsParam)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    byID, err := h.repo.FindByIDs(ctx, ids)
    if err != nil {
        http.Error(w, "internal", http.StatusInternalServerError)
        return
    }
    _ = json.NewEncoder(w).Encode(byID)
    return
}
```

### Step 4: イテレーション順序を観察

```go
func TestMap_IterationOrderIsRandom(t *testing.T) {
    m := map[int]string{1: "a", 2: "b", 3: "c", 4: "d"}
    var seen []int
    for k := range m {
        seen = append(seen, k)
    }
    t.Logf("実行ごとに違う順序: %v", seen)
}
```

`go test -count=5 -v ./pkg/handler/...` で順序が変わることを観察。**順序を保証したいなら slice にしてソートする**。

### Step 5: 存在チェックの 2 値受け

```go
if p, ok := m[1]; ok {
    fmt.Println("found:", p.Title)
} else {
    fmt.Println("not found")
}
```

`m[1]` だけだとゼロ値 (`nil`) が返るので「存在しない」のか「存在するが nil」か区別できない。`ok` 受けを癖に。

## 完了条件

- `curl 'localhost:8080/products?ids=1,2,3'` で `{"1": {...}, "2": {...}}` 形式のレスポンス。
- ent のクエリログ (`ent.Debug()`) が **1 本だけ** (N+1 でない)。
- `?ids=1,1,2` で重複なく返る。

## 詰まりやすいポイント

- **`var m map[K]V` は nil**: `make` か `map[K]V{}` で初期化を忘れない。
- **map のキーは比較可能型のみ**: slice / map / function はキーにできない。
- **値が struct の時、直接更新できない**: `m[k].Field = x` はコンパイルエラー。値をポインタにするか取り出して再代入。

## 関連パターン

- `map[K]struct{}` を Set として使う
- Preallocate maps with size hint
- N+1 回避 (`WHERE IN ?`)
