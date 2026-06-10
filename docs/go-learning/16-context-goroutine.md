# 16. context × goroutine (errgroup)

## このお題で足す機能

トピック 13 で書いた `DashboardUsecase` を `golang.org/x/sync/errgroup` で書き直す。1 つ失敗したら他の goroutine も ctx 経由で即座にキャンセル。クライアント切断時の自動キャンセルも動く。

## ゴール

- `errgroup.WithContext(ctx)` で **共有 ctx + エラー集約** を 1 つの構造で扱う。
- goroutine リーク・race・先行キャンセルを同時に解決する慣用句。
- `errgroup.Group.SetLimit(n)` で並列数制限。

## TS との違い

```ts
// Promise.all: 1 つでも reject すると全体 reject。だが他は走り続ける
await Promise.all([p1, p2, p3]);
```

```go
// errgroup: 1 つ失敗で ctx キャンセル、他の goroutine も止まる
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return f1(ctx) })
g.Go(func() error { return f2(ctx) })
err := g.Wait()
```

## ステップ

### Step 1: 依存追加

```bash
go get golang.org/x/sync/errgroup
```

### Step 2: DashboardUsecase を書き直す

```go
package usecase

import (
    "context"
    "golang.org/x/sync/errgroup"
)

func (u *DashboardUsecase) Get(ctx context.Context) (*DashboardData, error) {
    var (
        sales    int
        lowStock int
        recent   []*entity.History
    )

    g, ctx := errgroup.WithContext(ctx)

    g.Go(func() error {
        v, err := u.history.SumAmount(ctx)
        if err != nil { return fmt.Errorf("sales: %w", err) }
        sales = v
        return nil
    })

    g.Go(func() error {
        v, err := u.product.CountOutOfStock(ctx)
        if err != nil { return fmt.Errorf("lowStock: %w", err) }
        lowStock = v
        return nil
    })

    g.Go(func() error {
        v, err := u.history.RecentN(ctx, 5)
        if err != nil { return fmt.Errorf("recent: %w", err) }
        recent = v
        return nil
    })

    if err := g.Wait(); err != nil {
        return nil, err
    }
    return &DashboardData{TotalSales: sales, LowStockCount: lowStock, Recent: recent}, nil
}
```

トピック 13 の channel + select より **見通しが圧倒的に良い**。

ポイント:
- `errgroup.WithContext` が返す ctx を **各 goroutine に必ず渡す**。1 つ失敗で他が自動キャンセル。
- 各変数を別スロットで書くので Mutex 不要。
- `g.Wait()` は最初の non-nil エラーを返す。

### Step 3: 失敗時の伝播を観察

`u.history.SumAmount` がエラーを返すように仕掛ける:

```go
func (s *stubHistory) SumAmount(ctx context.Context) (int, error) {
    return 0, errors.New("boom")
}
```

このとき `u.product.CountOutOfStock` の中で `ctx.Done()` をチェックすると、すぐに `context.Canceled` で抜けるのが確認できる。

### Step 4: SetLimit で並列数制限

外部 API を 100 件呼びたいが、サーバ側に DoS したくない:

```go
g, ctx := errgroup.WithContext(ctx)
g.SetLimit(10) // 同時 10 goroutine まで

for _, url := range urls {
    url := url
    g.Go(func() error {
        return fetchOne(ctx, url)
    })
}
return g.Wait()
```

### Step 5: handler でクライアント切断を観察

```bash
curl --max-time 1 'localhost:8080/dashboard'
# サーバログ: "sales: context canceled" 等が出て即終了
```

クライアント切断 → `r.Context()` Done → errgroup の ctx Done → 全 goroutine が `ctx.Done()` でリターン → `g.Wait()` がエラーを返す。

## 完了条件

- `curl 'localhost:8080/dashboard'` が正常レスポンス、3 クエリが並行 (DB ログ確認)。
- 1 つ失敗させると、他の goroutine も即停止することがログで確認できる。
- `go test -race ./pkg/usecase/...` 緑。
- `curl --max-time 1` でクライアント切断時に DB クエリも中断。

## 詰まりやすいポイント

- **errgroup.WithContext の ctx を使わない**: 各 goroutine が外側の ctx を使うと、cancel が伝わらず leak。
- **g.Go の中で wg.Add とか書く**: 不要。errgroup が内部で管理。
- **return nil 忘れ**: nil を返さないとコンパイルエラー。
- **SetLimit のデッドロック**: SetLimit(0) は無制限。負値は panic。

## 関連パターン

- `errgroup.WithContext` で並行 + 早期キャンセル
- 共有 ctx を全 goroutine に渡す
- SetLimit で並列数制限
- 各 goroutine が別スロットに書いて Mutex 不要に
