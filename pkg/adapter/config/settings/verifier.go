// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package settings

import (
	"cmp"
)

// OutOfRangeError indicates that a Value was out of its acceptable
// range, either less than its minimum valid value or greater than its
// maximum valid value.
type OutOfRangeError[T cmp.Ordered] struct {
	Value        *T   // The actual out-of-range value
	LessThanMin  bool // true if and only if min boundary is violated
	InvalidRange bool // true if and only if min is greater than max
}

// Error implements error interface and returns a string reporting that
// minimum or maximum boundary value was not respected.
func (e *OutOfRangeError[T]) Error() string {
	switch {
	case e.InvalidRange:
		return "min is greater than max"
	case e.LessThanMin:
		return "value is less than min"
	default:
		return "value is greater than max"
	}
}

// VerifyRange verifies the given value ensuring that it is either nil
// or is within the provided minb/maxb boundary values, if the boundary
// values were given as non-nil values themselves. In case of a wrong
// value, in addition to the returned error, the value itself will be
// updated in order to take the minb or maxb value and fall in the
// acceptable range of values.
func VerifyRange[T cmp.Ordered](
	value **T, minb, maxb *T,
) *OutOfRangeError[T] {
	switch {
	case minb != nil && maxb != nil && (*minb) > (*maxb):
		return &OutOfRangeError[T]{InvalidRange: true}
	case (*value) == nil:
		return nil
	}
	switch v := **value; {
	case minb != nil && v < *minb:
		**value = *minb
		return &OutOfRangeError[T]{Value: &v, LessThanMin: true}
	case maxb != nil && v > *maxb:
		**value = *maxb
		return &OutOfRangeError[T]{Value: &v, LessThanMin: false}
	}
	return nil
}
