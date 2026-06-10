# 07. メソッドとレシーバー

## このお題で足す機能

`History` の Entity メソッド (`IsRefundable`, `Total`) を追加し、`User` に `RecentPurchases` メソッドを足す。ドメインロジックを Entity に集約し、handler / usecase からビジネスルールが消える。

## ゴール

- 値レシーバとポインタレシーバの選び方の **判断軸** を持つ (mutate するか / 構造体が大きいか / Mutex を含むか)。
- **embedding (型埋め込み)** で構成的に機能を再利用する。
- インタフェースとの相互作用 (どっちのレシーバで定義してもインタフェースを満たすか) を理解。

## TS との違い

```ts
class History {
  isRefundable(now: Date): boolean { return ... }
  total(): number { return this.amount * this.unitPrice; }
}
```

```go
type History struct { CreatedAt time.Time; Amount, UnitPrice int }
func (h History) IsRefundable(now time.Time) bool { ... } // クラス本体に書かなくていい
func (h History) Total() int { return h.Amount * h.UnitPrice }
```

メソッドは型のブロック内に書く必要はなく、**同じパッケージ内**ならどこでも宣言可。

## ステップ

### Step 1: History にドメインメソッドを追加

`pkg/entity/history.go`:

```go
package entity

import "time"

type History struct {
    ID        HistoryID
    ProductID ProductID
    UserID    UserID
    Amount    int
    UnitPrice int
    CreatedAt time.Time
}

const refundWindow = 7 * 24 * time.Hour

// 値レシーバ: 状態を変えない読み取り系
func (h History) IsRefundable(now time.Time) bool {
    return now.Sub(h.CreatedAt) <= refundWindow
}

func (h History) Total() int {
    return h.Amount * h.UnitPrice
}

// ポインタレシーバ: 状態を変える書き込み系
func (h *History) MarkRefunded(now time.Time) error {
    if !h.IsRefundable(now) {
        return ErrRefundExpired
    }
    h.Amount = 0
    return nil
}
```

実務では**全てポインタレシーバに揃える**派閥が多い。一貫性が大事。

### Step 2: User に集約メソッド

```go
type User struct {
    ID        UserID
    Name      string
    Histories []*History
}

func (u *User) RecentPurchases(within time.Duration, now time.Time) []*History {
    out := make([]*History, 0, len(u.Histories))
    for _, h := range u.Histories {
        if now.Sub(h.CreatedAt) <= within {
            out = append(out, h)
        }
    }
    return out
}

func (u *User) TotalSpent() int {
    sum := 0
    for _, h := range u.Histories {
        sum += h.Total()
    }
    return sum
}
```

### Step 3: Embedding で監査メタデータを足す

```go
type Audited struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (a Audited) Age(now time.Time) time.Duration { return now.Sub(a.CreatedAt) }

type Product struct {
    Audited // ← 埋め込み。Product は Age() を自動で持つ
    ID    ProductID
    Title string
    Stock int
}

// 利用側
p := &Product{Audited: Audited{CreatedAt: time.Now()}}
fmt.Println(p.Age(time.Now())) // Audited のメソッドが呼べる
```

埋め込みは TS の `extends` に近いが、**継承ではなく構成**。同名メソッドが両方にあると **アウター優先**。

### Step 4: handler から呼んで動作確認

`pkg/handler/history_handler.go`:

```go
func (h *HistoryHandler) handleRefundable(w http.ResponseWriter, r *http.Request) {
    histories, _ := h.repo.FindByUser(r.Context(), uid)
    now := time.Now()
    out := make([]*entity.History, 0, len(histories))
    for _, hh := range histories {
        if hh.IsRefundable(now) {
            out = append(out, hh)
        }
    }
    _ = json.NewEncoder(w).Encode(out)
}
```

handler に「日付差分計算」のロジックが**消える**ことを実感する。

## 完了条件

- `go test ./pkg/entity/...` で `IsRefundable` / `Total` / `RecentPurchases` の境界値テスト緑 (境界値: ちょうど 7 日、7 日 + 1 秒、6 日 23:59:59)。
- `curl 'localhost:8080/histories?refundable=true&user_id=1'` で直近 7 日以内のみ返る。
- handler / usecase に `time.Now().Sub(...)` 等の日付計算が**残っていない**。

## 詰まりやすいポイント

- **値レシーバの mutate は無効**: `func (h History) MarkRefunded()` 内で `h.Amount = 0` しても呼び出し元には反映されない。1 度わざと踏んで観察すると良い。
- **embedding は継承ではない**: `p.Audited.CreatedAt` でも `p.CreatedAt` でもアクセス可。型変換 `(Audited)(p)` のような書き方はできない。
- **Stringer 自動実装**: `func (p Product) String() string` を定義すると `fmt.Println(p)` で呼ばれる。

## 関連パターン

- Embedding for composition
- Domain logic を Entity に集約
- 値 vs ポインタレシーバの一貫性
