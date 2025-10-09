package fee

import (
	"testing"
	"time"
)

func TestFee(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		in        time.Time
		want      int
		expectErr bool
	}{
		{
			name:      "daytime_10:00",
			in:        time.Date(2022, time.February, 1, 8, 0, 0, 0, time.UTC),
			want:      1000,
			expectErr: false,
		},
		{
			name:      "midnight_22:00",
			in:        time.Date(2022, time.February, 1, 22, 0, 0, 0, time.UTC),
			want:      1200,
			expectErr: false,
		},
		{
			name:      "early_morning_5:00",
			in:        time.Date(2022, time.February, 1, 5, 0, 0, 0, time.UTC),
			want:      900,
			expectErr: false,
		},
		{
			name:      "err_2:00",
			in:        time.Date(2022, time.February, 1, 2, 0, 0, 0, time.UTC),
			want:      0,
			expectErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Fee(tt.in)
			if tt.expectErr {
				if err == nil {
					t.Error("want err")
				}
			} else {
				if err != nil {
					t.Error("not want err")
				}
			}

			if got != tt.want {
				t.Errorf("want = %d, but got = %d", tt.want, got)
			}
		})
	}
}
