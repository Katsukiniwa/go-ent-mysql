# 19. テスト

## このお題で足す機能

`PurchaseUsecase` にテーブル駆動テストでカバレッジ 80%+ を目指す。`testify/assert` を導入し、ent の SQLite in-memory DB で repository テストも書く。`httptest` で handler のエンドツーエンドテスト。

## ゴール

- テーブル駆動テストの基本形を書ける。
- `t.Run` でサブテスト、`t.Parallel()` で並列化。
- ent + SQLite で **DB 接続したまま** ユニットテストを書く。
- `httptest.NewRecorder` で handler を素手で叩く。

## TS との違い

```ts
// Jest: describe.each
describe.each([[1, 2, 3], [0, 0, 0]])("add", (a, b, e) => {
  it("...", () => expect(add(a, b)).toBe(e));
});
```

```go
// Go: 標準 testing のテーブル駆動
tests := []struct{ name string; a, b, expected int }{
    {"normal", 1, 2, 3},
    {"zero", 0, 0, 0},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        if got := add(tt.a, tt.b); got != tt.expected {
            t.Errorf("got=%d want=%d", got, tt.expected)
        }
    })
}
```

## ステップ

### Step 1: PurchaseUsecase のテーブル駆動テスト

```go
func TestPurchase(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name      string
        stock     int
        amount    int
        wantStock int
        wantErr   error
    }{
        {"normal", 10, 3, 7, nil},
        {"exact", 5, 5, 0, nil},
        {"oneShort", 1, 2, 1, entity.ErrOutOfStock},
        {"zero", 1, 0, 1, &entity.ValidationError{}},
        {"negative", 1, -1, 1, &entity.ValidationError{}},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            pr := &stubProductRepo{products: map[entity.ProductID]*entity.Product{
                1: {ID: 1, Stock: tt.stock},
            }}
            hr := &stubHistoryRepo{}
            uc := NewPurchaseUsecase(pr, hr)

            err := uc.Purchase(context.Background(), 1, 100, tt.amount)

            switch want := tt.wantErr.(type) {
            case nil:
                if err != nil { t.Fatalf("err=%v", err) }
            case *entity.ValidationError:
                var ve *entity.ValidationError
                if !errors.As(err, &ve) { t.Fatalf("want ValidationError, got %v", err) }
            default:
                if !errors.Is(err, want) { t.Fatalf("want %v, got %v", want, err) }
            }
            if pr.products[1].Stock != tt.wantStock {
                t.Errorf("stock=%d want=%d", pr.products[1].Stock, tt.wantStock)
            }
        })
    }
}
```

ポイント:
- `tt := tt` (Go 1.22+ なら不要だが意図を明示)。
- `t.Parallel()` を **外側** と **t.Run の中** の両方で呼ぶとテストケース間で並列実行。

### Step 2: testify で簡潔に

```bash
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/require
```

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFoo(t *testing.T) {
    got, err := DoSomething()
    require.NoError(t, err) // 失敗で即終了
    assert.Equal(t, want, got)
    assert.Len(t, got.Items, 3)
}
```

`require` は失敗で即停止、`assert` は記録して継続。

### Step 3: ent + SQLite で repository をテスト

```go
import (
    _ "github.com/mattn/go-sqlite3"
    "github.com/katsukiniwa/go-ent-mysql/product/ent/enttest"
)

func newTestClient(t *testing.T) *ent.Client {
    return enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
}

func TestProductRepository_FindByID(t *testing.T) {
    t.Parallel()
    client := newTestClient(t)
    repo := NewProductRepository(client)

    p, _ := client.Product.Create().SetTitle("apple").SetStock(5).Save(context.Background())

    got, err := repo.FindByID(context.Background(), entity.ProductID(p.ID))
    require.NoError(t, err)
    assert.Equal(t, "apple", got.Title)
    assert.Equal(t, 5, got.Stock)
}
```

`enttest.Open` は SQLite in-memory を **テストごとに新規 DB** で起動。本物の MySQL は要らない。

### Step 4: httptest で handler テスト

```go
func TestPurchaseHandler(t *testing.T) {
    t.Parallel()
    client := newTestClient(t)
    productRepo := repository.NewProductRepository(client)
    historyRepo := repository.NewHistoryRepository(client)
    uc := usecase.NewPurchaseUsecase(productRepo, historyRepo)
    h := handler.NewPurchaseHandler(uc)

    _, _ = client.Product.Create().SetTitle("x").SetStock(10).Save(context.Background())

    req := httptest.NewRequest("POST", "/products/1/purchase?amount=3&user_id=1", nil)
    w := httptest.NewRecorder()
    h.ServeHTTP(w, req)

    assert.Equal(t, http.StatusNoContent, w.Code)

    p, _ := client.Product.Get(context.Background(), 1)
    assert.Equal(t, 7, p.Stock)
}
```

DB → repository → usecase → handler を全部通すエンドツーエンドテスト。

### Step 5: カバレッジを計測

```bash
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out -o cover.html
open cover.html
```

未テスト行が赤くハイライト。80% を目標に。

### Step 6: ベンチマーク (おまけ)

```go
func BenchmarkPurchase(b *testing.B) {
    uc := setupUC()
    for i := 0; i < b.N; i++ {
        _ = uc.Purchase(context.Background(), 1, 100, 1)
    }
}
```

```bash
go test -bench=. -benchmem ./pkg/usecase/...
```

`-benchmem` で alloc 数も見える。

## 完了条件

- `go test ./...` 緑、`go test -race ./...` も緑。
- `go test -cover ./pkg/usecase/...` が **80%+**。
- handler テストが in-memory SQLite で動く (本物の MySQL 不要)。
- 異常系 (ErrOutOfStock, ValidationError, not found) もテストでカバー。

## 詰まりやすいポイント

- **テストの ctx に Background**: 本番では `r.Context()` だが、テストでは `context.Background()`。
- **subtests の Parallel**: `t.Parallel()` を外側だけ書くと並列にならない。サブテストの中でも書く。
- **テストの DB クリーンアップ**: `enttest.Open` は t に紐づいて自動 close されるが、共有 DB のときは `t.Cleanup` で明示。
- **gofmt 違反で test failed**: `gofmt -d ./...` を CI に。`go vet ./...` も。

## 関連パターン

- テーブル駆動テスト
- `t.Parallel()` で並列化
- `testify/assert` + `testify/require`
- `enttest` で in-memory DB
- `httptest` で handler を素手で叩く
