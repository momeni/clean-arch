// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import (
	"fmt"
	"strconv"
	"strings"
)

// SemVer represents a released semantic version, consisting of three
// components. First component indicates the major version. Incrementing
// it represents backward-incompatible changes. Second component is the
// minor version which represents feature additions and changes which
// are backward compatible and visible in the versioned API level.
// The last component is the patch version. It represents internal
// implementation changes which are invisible in the API level.
//
// No pre-release version is considered because unreleased versions are
// not supposed to be maintained and so do not need migration.
// If such a migration was required in a development machine, the source
// version may be migrated back to some released version by one binary
// and then migrated forward to another unreleased version by another
// binary (where they are seen as a yet to be evolved released version).
type SemVer [3]uint

// UnmarshalText deserializes text byte slice as a string consisting of
// three dot-separated numbers and fills the sv SemVer instance. In case
// of errors, sv will be left unchanged.
func (sv *SemVer) UnmarshalText(text []byte) (err error) {
	p := strings.Split(string(text), ".")
	l := len(p)
	if l == 0 || l > 3 {
		return fmt.Errorf("the %q has wrong number of components", text)
	}
	var v [3]int
	for i := 0; i < l; i++ {
		v[i], err = strconv.Atoi(p[i])
		if err != nil {
			return fmt.Errorf("the %q component is not numeric", p[i])
		}
		if v[i] < 0 {
			return fmt.Errorf("the %q component is negative", p[i])
		}
	}
	(*sv)[0] = uint(v[0])
	(*sv)[1] = uint(v[1])
	(*sv)[2] = uint(v[2])
	return nil
}

// Marshal serializes sv semantic version as its string representation.
// This is required for YAML serialization.
func (sv *SemVer) Marshal() string {
	return sv.String()
}

// MarshalText implements encoding.TextMarshaler interface and
// serializes `sv` semantic version as its string representation.
func (sv *SemVer) MarshalText() ([]byte, error) {
	return []byte(sv.String()), nil
}

// String returns the sv semantic version as a dot-separated string
// consisting of three numbers like major.minor.patch where all numbers
// are non-negative.
func (sv SemVer) String() string {
	return fmt.Sprintf("%d.%d.%d", sv[0], sv[1], sv[2])
}
