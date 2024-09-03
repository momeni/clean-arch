// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package settings

import (
	"errors"
	"log/slog"
	"strings"
	"time"
)

// Duration is a specialization of the time.Duration which produces a
// more human-readable representation when marshaled using its Marshal
// method.
type Duration time.Duration

// UnmarshalText reifies the encoding.TextUnmarshaler interface, so
// a byte slice (e.g., read from a YAML file) can be decoded as a
// time duration. The format of the `data` argument should conform
// to the time.ParseDuration expected format. In absence of errors,
// a nil error will be returned and only then, `d` receiver will be
// updated to contain the decoded duration.
func (d *Duration) UnmarshalText(data []byte) error {
	dd, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*d = Duration(dd)
	return nil
}

// Marshal returns a string representation of the `d` time duration.
// If d is nil, nil will be returned, so it can be used by higher-level
// Marshal methods for creation of an alternative struct (to be encoded
// to YAML instead of the actual data types with help of a top-level
// MarshalYAML method).
// If d is not nil, it will be encoded as a string according to the
// time.Duration string representation format, e.g., 2h3m4s, with this
// difference that zero trailing values will be ignored. That is, no
// 0s or 0m0s suffix may be included, for sake of more readability.
// A zero time duration will be encoded as 0h.
// The returned pointer will refer to a newly allocated string variable.
//
// See pkg/adapter/config/cfg2.*Config.Marshal for an example usage.
func (d *Duration) Marshal() *string {
	if d == nil {
		return nil
	}
	s := (*time.Duration)(d).String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return &s
}

// MarshalText implements encoding.TextMarshaler interface and
// serializes `d` duration using its Marshal method.
// This interface is required for json serialization.
func (d *Duration) MarshalText() ([]byte, error) {
	if s := d.Marshal(); s != nil {
		return []byte(*s), nil
	}
	return nil, errors.New("nil duration")
}

// LogValue implements slog.LogValuer and returns a DurationValue if
// this Duration is not nil, otherwise, it returns a StringValue with
// the constant "nil-duration" value.
func (d *Duration) LogValue() slog.Value {
	if d == nil {
		return slog.StringValue("nil-duration")
	}
	return slog.DurationValue(time.Duration(*d))
}
