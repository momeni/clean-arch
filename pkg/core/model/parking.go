package model

import (
	"errors"
	"fmt"
)

type ParkingMode int

const (
	ParkingModeInvalid ParkingMode = iota
	ParkingModeOld
	ParkingModeNew
)

// ErrUnknownParkingMode indicates that a given string may not be parsed
// as a valid/known parking mode. This error encodes a description err
// string and does not communicate the invalid parking mode string
// itself because the caller of Parse already knows about the invalid
// parking mode string.
// An error should be devised with this assumption that caller is aware
// of the function which is returning that error in addition to its
// arguments and other relevant system states which may be known before
// calling the function which is returning the error.
// Thereafter, the caller should wrap the obtained error and add the
// function name and arguments (or alternative information which makes
// the error complete in its new context), so it can be returned.
// Ultimately, one caller which is responsible to consume the error,
// can determine the entire call stack information from the single
// error chain with no reflection.
var ErrUnknownParkingMode = errors.New("unknown parking mode")

// ParkingModeError indicates an invalid parking mode. This error
// contains the invalid mode as an integer. Principally, this error
// type is not required (read the doc of ErrUnknownParkingMode).
// However, it is declared in order to show how extra parameters may be
// included in an error. The rare scenario which requires such an error
// instances (with parameters) belongs to functions which find out about
// the parameter during their execution and not by their arguments.
// For example, if a range-loop index is relevant for an error, it may
// be wrapped and returned by error like this.
type ParkingModeError int

func (e ParkingModeError) Error() string {
	return fmt.Sprintf("invalid parking mode: %d", e)
}

func (p ParkingMode) Validate() error {
	switch p {
	case ParkingModeOld, ParkingModeNew:
		return nil
	default:
		return ParkingModeError(p)
	}
}

func (p ParkingMode) String() string {
	switch p {
	case ParkingModeOld:
		return "old"
	case ParkingModeNew:
		return "new"
	default:
		panic(ParkingModeError(p))
	}
}

func ParseParkingMode(p string) (ParkingMode, error) {
	switch p {
	case "old":
		return ParkingModeOld, nil
	case "new":
		return ParkingModeNew, nil
	default:
		return ParkingModeInvalid, ErrUnknownParkingMode
	}
}
