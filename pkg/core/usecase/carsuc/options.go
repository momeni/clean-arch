package carsuc

import (
	"errors"
	"fmt"
	"time"
)

// Option is a functional option for the cars use case.
type Option func(uc *UseCase) error

// WithOldParkingMethodDelay option configures a cars UseCase instance
// in order to incur as much as the given delay during the old-method
// parking operations. This option may be passed to the New() function.
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
