# 05. interface

## このお題で足す機能

Repository を interface 化し、`pkg/usecase/` 層を新設する。handler が ent client に直接触れない構造に切り替え、テスト時はモック実装を差し込めるようにする。

## ゴール

- Go の **暗黙実装** に慣れる (`implements` キーワード無し)。
- **Accept interfaces, return structs** の原則を実装で味わう。
- インタフェースは **使う側** で定義する慣習を知る。
- コンパイル時に「インタフェースを満たしているか」チェックする書き方を覚える。

## TS との違い

```ts
class ProductRepo implements IProductRepo { ... } // 明示
```

```go
type ProductRepo struct{}
func (r *ProductRepo) FindByID(ctx context.Context, id int) (*Product, error) { ... }
// IProductRepo を満たしているかは「使う場所」で型検査される
```

## ステップ

### Step 1: Usecase 層を新設し、interface も usecase 側で定義

`pkg/usecase/purchase_usecase.go`:

```go
package usecase

import (
    "context"
    "github.com/katsukiniwa/go-ent-mysql/product/pkg/entity"
)

// 使う側 (usecase) で interface を宣言する
type ProductRepository interface {
    FindByID(ctx context.Context, id entity.ProductID) (*entity.Product, error)
    Save(ctx context.Context, p *entity.Product) error
}

type HistoryRepository interface {
    Create(ctx context.Context, h *entity.History) error
}

type PurchaseUsecase struct {
    products  ProductRepository
    histories HistoryRepository
}

func NewPurchaseUsecase(p ProductRepository, h HistoryRepository) *PurchaseUsecase {
    return &PurchaseUsecase{products: p, histories: h}
}

func (u *PurchaseUsecase) Purchase(ctx context.Context, pid entity.ProductID, uid entity.UserID, amount int) error {
    p, err := u.products.FindByID(ctx, pid)
    if err != nil {
        return err
    }
    if err := p.DecrementStock(amount); err != nil {
        return err
    }
    if err := u.products.Save(ctx, p); err != nil {
        return err
    }
    return u.histories.Create(ctx, &entity.History{
        ProductID: pid, UserID: uid, Amount: amount,
    })
}
```

**返り値は具象型 `*PurchaseUsecase`**、引数は **interface**。これが "Accept interfaces, return structs"。

### Step 2: 既存 Repository を interface に適合

`pkg/infrastructure/repository/product/repository.go` の `FindByID` / `Save` を usecase のシグネチャに合わせる。**コンパイル時チェック**を末尾に追加:

```go
// usecase.ProductRepository を満たすことをコンパイル時に検証
var _ usecase.ProductRepository = (*ProductRepository)(nil)
```

メソッドが足りなければ `go build` が落ちる。これが Go の "implements" 代替。

### Step 3: handler を usecase 経由に切り替え

`pkg/handler/purchase_handler.go`:

```go
type PurchaseHandler struct {
    uc *usecase.PurchaseUsecase
}

func NewPurchaseHandler(uc *usecase.PurchaseUsecase) *PurchaseHandler {
    return &PurchaseHandler{uc: uc}
}

func (h *PurchaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    pid, _ := entity.ParseProductID(r.URL.Query().Get("id"))
    uid, _ := entity.ParseUserID(r.URL.Query().Get("user_id"))
    amount, _ := strconv.Atoi(r.URL.Query().Get("amount"))

    if err := h.uc.Purchase(ctx, pid, uid, amount); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
```

`main.go`:

```go
purchaseUC := usecase.NewPurchaseUsecase(productRepo, historyRepo)
purchaseHandler := handler.NewPurchaseHandler(purchaseUC)
```

### Step 4: モックでテストを書く

`pkg/usecase/purchase_usecase_test.go`:

```go
type stubProductRepo struct {
    products map[entity.ProductID]*entity.Product
    saved    []*entity.Product
}

func (s *stubProductRepo) FindByID(_ context.Context, id entity.ProductID) (*entity.Product, error) {
    p, ok := s.products[id]
    if !ok {
        return nil, errors.New("not found")
    }
    return p, nil
}
func (s *stubProductRepo) Save(_ context.Context, p *entity.Product) error {
    s.saved = append(s.saved, p)
    return nil
}

type stubHistoryRepo struct{ created []*entity.History }

func (s *stubHistoryRepo) Create(_ context.Context, h *entity.History) error {
    s.created = append(s.created, h)
    return nil
}

func TestPurchase_DecrementsAndRecords(t *testing.T) {
    pr := &stubProductRepo{products: map[entity.ProductID]*entity.Product{
        1: {ID: 1, Stock: 10},
    }}
    hr := &stubHistoryRepo{}
    uc := NewPurchaseUsecase(pr, hr)

    if err := uc.Purchase(context.Background(), 1, 100, 3); err != nil {
        t.Fatal(err)
    }
    if pr.saved[0].Stock != 7 {
        t.Fatalf("stock=%d", pr.saved[0].Stock)
    }
    if len(hr.created) != 1 {
        t.Fatal("history が作られていない")
    }
}
```

**DB 接続なしでユニットテストが回る**。これが interface 抽出の最大のご褒美。

## 完了条件

- `go build ./...` 緑。
- `go test ./pkg/usecase/...` が DB 不要で通る。
- `var _ usecase.ProductRepository = (*ProductRepository)(nil)` が効いている (メソッド名をわざと typo してビルド失敗を観察)。

## 詰まりやすいポイント

- **大きすぎる interface**: 20 メソッド入れたくなるが、使う側が必要なメソッドだけ持つ**小さな interface** が良い。`Reader`/`Writer` の哲学。
- **interface を返す関数**: 原則として具象型を返す。返した側が後で機能追加した時、interface 経由だと使えない。
- **`any` (= `interface{}`) の濫用**: コンパイラの型チェックが効かなくなる。最終手段。

## 関連パターン

- Accept interfaces, return structs
- Define interfaces where they're used (consumer-side)
- Compile-time satisfaction check: `var _ I = (*T)(nil)`
