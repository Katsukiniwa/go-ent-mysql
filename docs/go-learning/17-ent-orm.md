# 17. ent (ORM)

## このお題で足す機能

新しいスキーマ `Cart` / `CartItem` を追加。`User ↔ Cart ↔ Product` のリレーションを定義し、`GET /users/{id}/cart` と `POST /users/{id}/cart/items` を実装。`With...` で N+1 を回避。

## ゴール

- ent でスキーマを足す → `go generate` → 自動生成コードを使う流れを習得。
- `Edges` (`edge.To` / `edge.From`) でリレーションを宣言。
- `Query().With...()` で eager loading。
- ent のトランザクション API の慣用句。

## TS との違い

ent は **コード生成型 ORM** (Prisma に近い)。スキーマ → 型安全なクエリビルダ生成。生成されたコードは `ent.Product.Query().Where(...)` のように補完が効く。

## ステップ

### Step 1: Cart / CartItem スキーマを定義

`ent/schema/cart.go`:

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
)

type Cart struct{ ent.Schema }

func (Cart) Fields() []ent.Field {
    return []ent.Field{
        field.Int("user_id").Optional(),
    }
}

func (Cart) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).
            Ref("cart").
            Field("user_id").
            Unique(),
        edge.To("items", CartItem.Type),
    }
}
```

`ent/schema/cart_item.go`:

```go
type CartItem struct{ ent.Schema }

func (CartItem) Fields() []ent.Field {
    return []ent.Field{
        field.Int("cart_id").Optional(),
        field.Int("product_id").Optional(),
        field.Int("quantity").Default(1),
    }
}

func (CartItem) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("cart", Cart.Type).Ref("items").Field("cart_id").Unique(),
        edge.From("product", Product.Type).Ref("cart_items").Field("product_id").Unique(),
    }
}
```

`User` / `Product` の Edges も更新:

```go
// user.go
edge.To("cart", Cart.Type).Unique(),
// product.go
edge.To("cart_items", CartItem.Type),
```

### Step 2: コード生成

```bash
go generate ./ent
```

`ent/cart/`, `ent/cartitem/` 等のパッケージが生成される。

### Step 3: Repository を書く

`pkg/infrastructure/repository/cart/repository.go`:

```go
func (r *CartRepository) FindByUser(ctx context.Context, uid entity.UserID) (*entity.Cart, error) {
    c, err := r.client.Cart.Query().
        Where(cart.HasUserWith(user.IDEQ(int(uid)))).
        WithItems(func(q *ent.CartItemQuery) {
            q.WithProduct() // N+1 を避ける
        }).
        First(ctx)
    if err != nil {
        if ent.IsNotFound(err) {
            return nil, entity.ErrNotFound
        }
        return nil, fmt.Errorf("FindByUser(%d): %w", uid, err)
    }
    return toEntity(c), nil
}
```

`WithItems` + `WithProduct` で **JOIN を 1 クエリにまとめる**。`ent.Debug()` のログで 1 SQL に統合されていることを確認。

### Step 4: AddItem (Upsert パターン)

```go
func (r *CartRepository) AddItem(ctx context.Context, uid entity.UserID, pid entity.ProductID, qty int) error {
    return product.WithTx(ctx, r.client, func(tx *ent.Tx) error {
        c, err := tx.Cart.Query().
            Where(cart.HasUserWith(user.IDEQ(int(uid)))).First(ctx)
        if ent.IsNotFound(err) {
            c, err = tx.Cart.Create().SetUserID(int(uid)).Save(ctx)
        }
        if err != nil { return err }

        existing, err := tx.CartItem.Query().
            Where(cartitem.CartID(c.ID), cartitem.ProductID(int(pid))).
            First(ctx)
        switch {
        case ent.IsNotFound(err):
            _, err = tx.CartItem.Create().
                SetCartID(c.ID).
                SetProductID(int(pid)).
                SetQuantity(qty).Save(ctx)
        case err != nil:
            return err
        default:
            _, err = tx.CartItem.UpdateOne(existing).
                SetQuantity(existing.Quantity + qty).Save(ctx)
        }
        return err
    })
}
```

### Step 5: handler

```go
func (h *CartHandler) get(w http.ResponseWriter, r *http.Request) {
    uid, _ := strconv.Atoi(r.PathValue("id"))
    c, err := h.uc.GetByUser(r.Context(), entity.UserID(uid))
    if err != nil { http.Error(w, "err", 500); return }
    _ = json.NewEncoder(w).Encode(c)
}

func (h *CartHandler) addItem(w http.ResponseWriter, r *http.Request) {
    uid, _ := strconv.Atoi(r.PathValue("id"))
    var body struct{ ProductID, Quantity int }
    json.NewDecoder(r.Body).Decode(&body)
    err := h.uc.AddItem(r.Context(), entity.UserID(uid), entity.ProductID(body.ProductID), body.Quantity)
    if err != nil { http.Error(w, "err", 500); return }
    w.WriteHeader(http.StatusNoContent)
}
```

### Step 6: マイグレーション

```bash
go run . # main.go の client.Schema.Create で auto migrate
```

本番では `atlas` を使った差分マイグレーション推奨。学習段階では auto migrate で OK。

## 完了条件

- `curl -X POST 'localhost:8080/users/1/cart/items' -d '{"ProductID":1,"Quantity":2}'` で Cart に item が入る。
- `curl 'localhost:8080/users/1/cart'` で items 一覧 (product 情報込み) が返る。
- `ent.Debug()` のログで eager loading が **1 SQL** にまとまっていることを確認。
- 同じ product を 2 回 add すると quantity が加算される (Upsert)。

## 詰まりやすいポイント

- **Edge の inverse**: `edge.To` と `edge.From` はペア。片方だけ書くと整合性エラー。
- **go generate 忘れ**: スキーマを変えたら必ず `go generate ./ent`。CI でも実行。
- **N+1**: `WithItems` を付けず for ループで `c.QueryItems().All(ctx)` を呼ぶと N+1。`ent.Debug()` でクエリ数を見る癖を。
- **Unique 制約のエラー**: `ent.IsConstraintError(err)` で判別できる。

## 関連パターン

- ent の eager loading (`With...`)
- Repository + TX ヘルパで一連の更新を atomic に
- スキーマ変更 → コード生成 → ビルドの三段階
