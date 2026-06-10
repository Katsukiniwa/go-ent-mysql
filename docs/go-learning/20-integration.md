# 20. 総合演習: 予約フロー

## このお題で足す機能

`POST /reservations` を実装する。「Cart の内容を一括購入 → 全商品の在庫減算 + History 作成 + Cart を空にする」を **1 トランザクション** で行う。これまでのトピックを全部使う。

- 入力検証 (06)
- ctx 貫通 (10) + timeout (12)
- request-id ログ (11)
- errgroup で在庫一括チェック (16)
- トランザクション (08, 17)
- テーブル駆動テスト + httptest (19)
- Wire で配線 (18)

## ゴール

- ここまでの全 19 トピックの内容を 1 つの機能に統合できる。
- 障害ケース (在庫不足 / DB エラー / クライアント切断) で **整合性が保たれる** ことを確認。
- 自分の手で書ききった達成感を得る。

## ステップ

### Step 1: API 設計

```
POST /reservations
Headers: X-Request-ID: <uuid>
Body: { "user_id": 1 }

200 OK:
{ "reservation_id": 42, "total": 12000, "items": [...] }

400: ユーザー ID 不正 / Cart 空
404: User not found
409: いずれかの商品が在庫不足
504: タイムアウト
500: 内部エラー
```

### Step 2: Reservation スキーマを追加

`ent/schema/reservation.go`:

```go
type Reservation struct{ ent.Schema }

func (Reservation) Fields() []ent.Field {
    return []ent.Field{
        field.Int("user_id"),
        field.Int("total"),
        field.Time("created_at").Default(time.Now),
    }
}

func (Reservation) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("user", User.Type).Ref("reservations").Field("user_id").Unique(),
        edge.To("histories", History.Type),
    }
}
```

`User` 側に `edge.To("reservations", Reservation.Type)` も追加。

```bash
go generate ./ent
```

### Step 3: Usecase

`pkg/usecase/reservation_usecase.go`:

```go
type ReservationUsecase struct {
    client *ent.Client
    log    *slog.Logger
}

type ReservationResult struct {
    ReservationID int
    Total         int
    Items         []ReservationLine
}
type ReservationLine struct {
    ProductID entity.ProductID
    Quantity  int
    Subtotal  int
}

func (u *ReservationUsecase) Create(ctx context.Context, uid entity.UserID) (*ReservationResult, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    u.log.InfoContext(ctx, "reservation start", "user_id", uid)

    var result *ReservationResult
    err := product.WithTx(ctx, u.client, func(tx *ent.Tx) error {
        // 1. Cart を取得
        c, err := tx.Cart.Query().
            Where(cart.HasUserWith(user.IDEQ(int(uid)))).
            WithItems(func(q *ent.CartItemQuery) { q.WithProduct() }).
            Only(ctx)
        if ent.IsNotFound(err) { return entity.ErrUserNotFound }
        if err != nil { return fmt.Errorf("query cart: %w", err) }
        if len(c.Edges.Items) == 0 {
            return &entity.ValidationError{Field: "cart", Message: "empty"}
        }

        // 2. 在庫を errgroup で並行チェック
        g, gctx := errgroup.WithContext(ctx)
        for _, item := range c.Edges.Items {
            item := item
            g.Go(func() error {
                p, err := tx.Product.Get(gctx, item.Edges.Product.ID)
                if err != nil { return err }
                if p.Stock < item.Quantity {
                    return fmt.Errorf("product %d: %w", p.ID, entity.ErrOutOfStock)
                }
                return nil
            })
        }
        if err := g.Wait(); err != nil { return err }

        // 3. 在庫減算 + 集計
        total := 0
        lines := make([]ReservationLine, 0, len(c.Edges.Items))
        for _, item := range c.Edges.Items {
            p := item.Edges.Product
            if _, err := tx.Product.UpdateOne(p).
                SetStock(p.Stock - item.Quantity).Save(ctx); err != nil {
                return fmt.Errorf("update stock %d: %w", p.ID, err)
            }
            subtotal := item.Quantity * 100 // 仮の単価
            total += subtotal
            lines = append(lines, ReservationLine{
                ProductID: entity.ProductID(p.ID),
                Quantity:  item.Quantity,
                Subtotal:  subtotal,
            })
        }

        // 4. Reservation 作成
        res, err := tx.Reservation.Create().
            SetUserID(int(uid)).
            SetTotal(total).Save(ctx)
        if err != nil { return fmt.Errorf("create reservation: %w", err) }

        // 5. History 作成
        for _, line := range lines {
            if _, err := tx.History.Create().
                SetUserID(int(uid)).
                SetAmount(line.Quantity).Save(ctx); err != nil {
                return fmt.Errorf("create history: %w", err)
            }
        }

        // 6. Cart を空にする
        if _, err := tx.CartItem.Delete().
            Where(cartitem.CartID(c.ID)).Exec(ctx); err != nil {
            return fmt.Errorf("clear cart: %w", err)
        }

        result = &ReservationResult{
            ReservationID: res.ID, Total: total, Items: lines,
        }
        return nil
    })

    if err != nil {
        u.log.ErrorContext(ctx, "reservation failed", "err", err)
        return nil, err
    }
    u.log.InfoContext(ctx, "reservation done", "id", result.ReservationID, "total", result.Total)
    return result, nil
}
```

### Step 4: Handler

```go
func (h *ReservationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    var body struct{ UserID int }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "bad request", 400); return
    }
    result, err := h.uc.Create(r.Context(), entity.UserID(body.UserID))

    var ve *entity.ValidationError
    switch {
    case err == nil:
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(result)
    case errors.As(err, &ve):
        http.Error(w, ve.Error(), 400)
    case errors.Is(err, entity.ErrUserNotFound):
        http.Error(w, "user not found", 404)
    case errors.Is(err, entity.ErrOutOfStock):
        http.Error(w, err.Error(), 409)
    case errors.Is(err, context.DeadlineExceeded):
        http.Error(w, "gateway timeout", 504)
    default:
        h.log.ErrorContext(r.Context(), "internal", "err", err)
        http.Error(w, "internal", 500)
    }
}
```

### Step 5: Wire に追加

```go
wire.Build(
    // ... 既存
    usecase.NewReservationUsecase,
    handler.NewReservationHandler,
)
```

```bash
cd cmd/server && wire
```

### Step 6: テスト

```go
func TestReservation_Create(t *testing.T) {
    t.Parallel()
    client := newTestClient(t)
    // setup user, products, cart items

    uc := NewReservationUsecase(client, slog.Default())
    got, err := uc.Create(context.Background(), entity.UserID(1))
    require.NoError(t, err)
    assert.Equal(t, 2, len(got.Items))

    p, _ := client.Product.Get(context.Background(), 1)
    assert.Equal(t, 7, p.Stock)
    n, _ := client.CartItem.Query().Where(cartitem.CartIDEQ(1)).Count(context.Background())
    assert.Equal(t, 0, n)
}

func TestReservation_OutOfStock_RollsBack(t *testing.T) {
    t.Parallel()
    client := newTestClient(t)
    // Cart に在庫を超える数量を入れる
    uc := NewReservationUsecase(client, slog.Default())
    _, err := uc.Create(context.Background(), entity.UserID(1))
    assert.ErrorIs(t, err, entity.ErrOutOfStock)

    // ロールバック確認
    p, _ := client.Product.Get(context.Background(), 1)
    assert.Equal(t, 10, p.Stock)
    n, _ := client.CartItem.Query().Where(cartitem.CartIDEQ(1)).Count(context.Background())
    assert.Greater(t, n, 0)
}
```

### Step 7: 実機で叩く

```bash
docker compose up -d
go run ./cmd/server

# 別ターミナル
curl -X POST 'localhost:8080/users/1/cart/items' -d '{"ProductID":1,"Quantity":3}'
curl -X POST 'localhost:8080/users/1/cart/items' -d '{"ProductID":2,"Quantity":1}'
curl -i -X POST 'localhost:8080/reservations' -H 'Content-Type: application/json' \
     -H 'X-Request-ID: demo-001' -d '{"user_id":1}'
```

サーバログに `request_id=demo-001` が貫通して `reservation done` が出れば完走。

## 完了条件

- 正常系: Reservation が作成され、Cart が空になり、Product.stock が減っている。
- 異常系 (在庫不足): 何も変わっていない (ロールバック)。
- タイムアウト: 504 が返る (DB を一時停止すれば再現可能)。
- ログに request_id が全層で貫通。
- `go test -race -cover ./...` 緑、カバレッジ 80%+。

## ここまで完走したあなたが得たもの

- 型・slice・map・interface・error・goroutine・channel・context・errgroup・defer・mutex・testing・ORM・DI を **実プロジェクトに足しながら**学べた。
- 「TS の Promise.all で書けるじゃん」→「Go の errgroup の方がエラー伝播が明示的で良いな」のような **言語を越えた比較感覚**が身に付いた。
- このリポジトリが小さな EC バックエンドに育った。

次は: gRPC, GraphQL, OpenTelemetry, k8s デプロイ あたりに進むと面白い。

## 関連パターン

- すべての関連パターンの集大成
- トランザクション境界の設計
- 観測可能性 (request-id, structured log)
- 失敗時のロールバック保証
