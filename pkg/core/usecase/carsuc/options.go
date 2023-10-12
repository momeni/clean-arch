package carsuc

import (
	"errors"
	"fmt"
	"time"
)

type Option func(uc *UseCase) error

func WithOldParkingMethodDelay(delay time.Duration) Option {
	return func(uc *UseCase) error {
		if d := int64(delay); d <= 0 {
			return fmt.Errorf("delay (%d) is not positive", d)
		}
		if uc.oldParkingMethodDelay != 0 {
			return errors.New("delay is already configured")
		}
		uc.oldParkingMethodDelay = delay
		return nil
	}
}
