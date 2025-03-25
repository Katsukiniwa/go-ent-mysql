package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/katsukiniwa/go-ent-mysql/product/ent"
	"github.com/katsukiniwa/go-ent-mysql/product/ent/history"
	"github.com/katsukiniwa/go-ent-mysql/product/ent/user"
	"github.com/katsukiniwa/go-ent-mysql/product/pkg/entity"
)

type HistoryRepository interface {
	GetHistories(ctx context.Context) ([]entity.History, error)
	InsertHistory(ctx context.Context, userId int, amount int) error
	UpdateHistory(ctx context.Context, id int, amount int) error
	DeleteHistory(ctx context.Context, id int) error
}

type historyRepository struct {
	client *ent.Client
}

func NewHistoryRepository(client *ent.Client) HistoryRepository {
	return &historyRepository{client: client}
}

func (pr *historyRepository) GetHistories(ctx context.Context) ([]entity.History, error) {
	histories, err := pr.client.History.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed querying histories: %w", err)
	}

	log.Println("histories: ", histories)

	result := make([]entity.History, 0, len(histories))

	for _, v := range histories {
		result = append(result, entity.History{ID: v.ID, UserID: v.ID, Amount: v.Amount})
	}

	return result, nil
}

func (pr *historyRepository) UpdateHistory(ctx context.Context, id int, amount int) error {
	_, err := pr.client.History.UpdateOneID(id).SetAmount(amount).Save(ctx)
	if err != nil {
		return fmt.Errorf("failed updating history: %w", err)
	}

	return nil
}

func (pr *historyRepository) DeleteHistory(ctx context.Context, id int) error {
	err := pr.client.History.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed updating history: %w", err)
	}

	return nil
}

func (pr *historyRepository) InsertHistory(ctx context.Context, userId int, amount int) error {
	if _, err := pr.client.ExecContext(ctx, "SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED"); err != nil {
		return fmt.Errorf("failed to set transaction isolation level: %w", err)
	}

	// 一日の最大出金金額を超えてないかチェックして出金履歴を登録するトランザクション開始
	tx, err := pr.client.Tx(ctx)
	if err != nil {
		log.Println("starting a transaction: %w", err)
		tx.Rollback()

		return fmt.Errorf("failed to start a transaction: %w", err)
	}

	// 事前にuserレコードのロックを取得する
	_, err = tx.Client().User.Query().ForUpdate().Where(user.ID(userId)).Only(ctx)
	if err != nil {
		log.Println("ユーザレコードのロック取得に失敗しました: %w", err)
		tx.Rollback()

		return fmt.Errorf("failed to get user record: %w", err)
	}

	// ユーザーの合計出金金額を取得
	var histories []struct {
		Sum int `json:"sum"`
	}

	err = tx.Client().History.Query().ForUpdate().Where(
		history.UserID(userId),
	).Aggregate(
		ent.Sum(history.FieldAmount),
	).
		Scan(ctx, &histories)
	if err != nil {
		log.Println("ユーザの出金履歴の取得に失敗しました: %w", err)
		tx.Rollback()

		return fmt.Errorf("failed to get user's history: %w", err)
	}

	ra := 0
	if len(histories) > 0 {
		ra = histories[0].Sum
	}

	// 最大出金金額を超えていたら400を返す
	if ra+amount > entity.AmountLimit {
		log.Println("最大出金金額を超えています")
		tx.Rollback()

		return nil
	}

	// 出金履歴登録
	if err := tx.History.Create().SetAmount(amount).SetUserID(userId).Exec(ctx); err != nil {
		log.Println("出金履歴の登録に失敗しました: %w", err)
		tx.Rollback()

		return fmt.Errorf("failed to insert history: %w", err)
	}

	// トランザクションコミット
	if err := tx.Commit(); err != nil {
		log.Println("ロールバックに失敗しました: %w", err)
		tx.Rollback()

		return fmt.Errorf("failed committing a transaction: %w", err)
	}

	return nil
}
