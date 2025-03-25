package fee

import (
	"errors"
	"time"
)

const defaultFee = 1000

func Fee(admissionTime time.Time) (int, error) {
	fee := float64(defaultFee)
	hour := admissionTime.Hour()
	eight := 8
	twentyTwo := 22

	switch {
	case hour >= 2 && hour < 5:
		return 0, errors.New("現在は入場できない時間帯です")
	case hour < eight:
		fee *= 0.9
	case hour >= twentyTwo:
		fee *= 1.2
	}

	return int(fee), nil
}
