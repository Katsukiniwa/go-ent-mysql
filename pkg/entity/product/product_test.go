package product

import (
	"testing"
	"time"
)

func TestProduct(t *testing.T) {
	cases := []struct {
		name      string
		in        time.Time
		histories []ProductPriceHistory
		want      int64
	}{
		{
			name: "EndedAtが存在しない価格履歴がある場合",
			in:   time.Date(2022, time.February, 15, 8, 0, 0, 0, time.UTC),
			histories: []ProductPriceHistory{
				{
					ProductID: 1,
					Price:     1000,
					StartedAt: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
					EndedAt:   time.Date(2022, time.January, 31, 0, 0, 0, 0, time.UTC),
				},
				{
					ProductID: 1,
					Price:     1200,
					StartedAt: time.Date(2022, time.February, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 1200,
		},
		{
			name: "StartedAtが存在しない価格履歴がある場合",
			in:   time.Date(2022, time.January, 15, 8, 0, 0, 0, time.UTC),
			histories: []ProductPriceHistory{
				{
					ProductID: 1,
					Price:     1000,
					EndedAt:   time.Date(2022, time.January, 31, 0, 0, 0, 0, time.UTC),
				},
				{
					ProductID: 1,
					Price:     1200,
					StartedAt: time.Date(2022, time.February, 1, 0, 0, 0, 0, time.UTC),
					EndedAt:   time.Date(2022, time.February, 29, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 1000,
		},
		{
			name: "全ての価格履歴のStartedAtとEndedAtが被っていない場合",
			in:   time.Date(2022, time.January, 15, 8, 0, 0, 0, time.UTC),
			histories: []ProductPriceHistory{
				{
					ProductID: 1,
					Price:     1000,
					StartedAt: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
					EndedAt:   time.Date(2022, time.January, 31, 0, 0, 0, 0, time.UTC),
				},
				{
					ProductID: 1,
					Price:     1200,
					StartedAt: time.Date(2022, time.February, 1, 0, 0, 0, 0, time.UTC),
					EndedAt:   time.Date(2022, time.February, 29, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 1000,
		},
		{
			name: "価格履歴の日付に漏れがあり2022年2月1日の価格が存在しない場合",
			in:   time.Date(2022, time.February, 1, 0, 0, 0, 0, time.UTC),
			histories: []ProductPriceHistory{
				{
					ProductID: 1,
					Price:     1000,
					StartedAt: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
					EndedAt:   time.Date(2022, time.January, 31, 0, 0, 0, 0, time.UTC),
				},
				{
					ProductID: 1,
					Price:     1200,
					StartedAt: time.Date(2022, time.February, 2, 0, 0, 0, 0, time.UTC),
					EndedAt:   time.Date(2022, time.February, 29, 0, 0, 0, 0, time.UTC),
				},
			},
			want: 0,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := &Product{
				PriceHistories: tt.histories,
			}

			if got.CurrentPrice(tt.in) != tt.want {
				t.Errorf("got: %d, want: %d", got.CurrentPrice(tt.in), tt.want)
			}
		})
	}
}
