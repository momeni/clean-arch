// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package settings

// Nil2Zero overwrites the (*t) pointer, which should be nil,
// in order to point to a newly allocated T instance and initializes it
// with the zero value of T type.
// If the (*t) pointer was not nil, Nil2Zero will perform no action.
func Nil2Zero[T any](t **T) {
	if (*t) != nil {
		return
	}
	var zero T
	(*t) = &zero
}

// OverwriteNil overwrites the (*dst) pointer, which should be nil,
// in order to point to a newly allocated T instance and initializes it
// with the (*src) value.
// If the (*dst) pointer was not nil or if the src was nil, this
// function will perform no action.
func OverwriteNil[T any](dst **T, src *T) {
	if (*dst) != nil || src == nil {
		return
	}
	t := *src
	(*dst) = &t
}

// OverwriteUnconditionally overwrites the (*dst) pointer, which may or
// may not be a nil pointer, unconditionally based on the src pointer.
// If the src pointer is nil, (*dst) is overwritten so it becomes nil.
// If the src pointer is not nil, (*dst) is overwritten so it points to
// a newly allocated T instance which is initialized by (*src) value.
func OverwriteUnconditionally[T any](dst **T, src *T) {
	if src == nil {
		(*dst) = nil
		return
	}
	t := *src
	(*dst) = &t
}
