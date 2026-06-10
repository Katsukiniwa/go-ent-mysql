# 08. defer

## このお題で足す機能

`Purchase` ユースケースに **DB トランザクション** を導入する。`defer tx.Rollback()` で確実にクリーンアップする慣用句を体得し、「在庫減算 + 履歴作成」を 1 トランザクションに収める。

## ゴール

- `defer` の LIFO 順を理解。
- `defer + recover` でパニックを業務に漏らさないパターンを書ける。
- ループ内 `defer` リーク (関数末尾まで保留される) を回避できる。
- ent のトランザクション API (`client.Tx(ctx)`) を使える。

## TS との違い

```ts
const conn = await pool.connect();
try {
  await conn.query("BEGIN");
  await conn.query("COMMIT");
} catch (e) {
  await conn.query("ROLLBACK"); throw e;
} finally { conn.release(); }
```

```go
tx, err := client.Tx(ctx)
if err != nil { return err }
defer tx.Rollback() // Commit 後の Rollback はエラーで返るだけで害なし

return tx.Commit()
```

## ステップ

### Step 1: トランザクションヘルパを書く

`pkg/infrastructure/repository/product/tx.go`:

```go
package product

import (
    "context"
    "fmt"

    "github.com/katsukiniwa/go-ent-mysql/product/ent"
)

func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) (err error) {
    tx, err := client.Tx(ctx)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    // panic が起きても Rollback して再 panic
    defer func() {
        if v := recover(); v != nil {
            _ = tx.Rollback()
            panic(v)
        }
    }()
    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("%w (rollback failed: %v)", err, rbErr)
        }
        return err
    }
    return tx.Commit()
}
```

ポイント:
- panic を `recover` して **Rollback してから再 panic**。業務エラーは `error` で扱い、想定外 panic はクリーンアップだけして上に投げる。
- 業務側は `WithTx(...)` のクロージャに集中できる。

### Step 2: Usecase でトランザクションに包む

```go
func (u *PurchaseUsecase) Purchase(ctx context.Context, pid entity.ProductID, uid entity.UserID, amount int) error {
    return product.WithTx(ctx, u.client, func(tx *ent.Tx) error {
        p, err := tx.Product.Get(ctx, int(pid))
        if err != nil {
            return wrapNotFound(err, pid)
        }
        if p.Stock < amount {
            return entity.ErrOutOfStock
        }

        if _, err := tx.Product.UpdateOne(p).
            SetStock(p.Stock - amount).Save(ctx); err != nil {
            return fmt.Errorf("update stock: %w", err)
        }
        if _, err := tx.History.Create().
            SetUserID(int(uid)).
            SetAmount(amount).Save(ctx); err != nil {
            return fmt.Errorf("create history: %w", err)
        }
        return nil
    })
}
```

**1 つのトランザクション内**で在庫減算+履歴作成が atomic に。途中失敗で全体ロールバック。

### Step 3: defer の LIFO を観察

`pkg/usecase/defer_demo_test.go`:

```go
func TestDefer_LIFO(t *testing.T) {
    var got []string
    log := func(s string) { got = append(got, s) }

    func() {
        defer log("first")
        defer log("second")
        defer log("third")
        log("body")
    }()

    want := []string{"body", "third", "second", "first"}
    for i := range want {
        if got[i] != want[i] {
            t.Fatalf("got=%v", got)
        }
    }
}
```

defer は **後入れ先出し** (積み下ろし)。

### Step 4: ループ内 defer のアンチパターン

```go
// Bad: 1000 ファイルの Close が関数末尾まで保留 (FD リーク risk)
func processBad(paths []string) error {
    for _, p := range paths {
        f, err := os.Open(p)
        if err != nil { return err }
        defer f.Close() // 1000 個積み上がる
    }
    return nil
}

// Good: イテレーションごとに別スコープへ
func processGood(paths []string) error {
    for _, p := range paths {
        if err := func() error {
            f, err := os.Open(p)
            if err != nil { return err }
            defer f.Close()
            // ... process f
            return nil
        }(); err != nil {
            return err
        }
    }
    return nil
}
```

## 完了条件

- 在庫不足ケースで `SELECT stock FROM products WHERE id=1` の値が**変わっていない**。
- 正常購入で history が +1、product.stock が `amount` 減。
- `go test -run TestDefer_LIFO ./pkg/usecase/...` 緑。

## 詰まりやすいポイント

- **defer の引数評価は登録時**: `defer log(time.Now())` は登録時刻が記録される。実行時にしたいなら `defer func() { log(time.Now()) }()`。
- **`os.Exit` / `log.Fatal` では defer 実行されない**。panic では実行される。
- **named return + defer で戻り値書き換え**: `func() (err error) { defer func(){ err = fmt.Errorf("wrap: %w", err) }(); ... }` 便利だが読み手の認知負荷が上がるので多用しない。

## 関連パターン

- `defer + recover` で panic から service を守る
- ループ内 defer は別関数化
- Cleanup (FD / lock / tx) は必ず defer
