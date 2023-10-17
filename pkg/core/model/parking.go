// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"errors"
	"fmt"
)

// ParkingMode specifies the parking mode enum and accepts two
// old and new methods. Although this enum is numeric, it is
// (de)serialized as a string for readability in the adapter layer.
type ParkingMode int

// Valid values for the ParkingMode enum.
const (
	ParkingModeInvalid ParkingMode = iota // zero value is invalid

	ParkingModeOld // old method incurs more delay
	ParkingModeNew // new method parks the car with no delay
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

// Error implements the error interface, returning a string
// representation of the ParkingModeError.
func (e ParkingModeError) Error() string {
	return fmt.Sprintf("invalid parking mode: %d", e)
}

// Validate returns nil if ParkingMode value is valid. For invalid
// values, an instance of the ParkingModeError will be returned.
func (p ParkingMode) Validate() error {
	switch p {
	case ParkingModeOld, ParkingModeNew:
		return nil
	default:
		return ParkingModeError(p)
	}
}

// String converts the ParkingMode enum to a string, helping to
// serialize it for transmission to web clients (for improved
// readability). Invalid parking mode causes a panic.
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

// ParseParkingMode parses the given string and returns a ParkingMode,
// helping to deserialize it when reading a REST API request.
// For invalid strings, ParkingModeInvalid and ErrUnknownParkingMode
// will be returned.
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
