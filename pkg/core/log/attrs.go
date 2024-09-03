// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package log

import (
	"log/slog"
)

// Valuer returns an Attr for the given slog.LogValuer value.
func Valuer(key string, value slog.LogValuer) slog.Attr {
	return slog.Any(key, value)
}

// Err returns an Attr for the given error value.
// The error value is resolved as a string by its Error() method.
// If error value is nil, the constant "no-error" value will be used.
func Err(key string, value error) slog.Attr {
	if value == nil {
		return slog.String(key, "no-error")
	}
	return slog.String(key, value.Error())
}
