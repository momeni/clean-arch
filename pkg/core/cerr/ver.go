// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cerr

import (
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/model"
)

// MismatchingSemVerError indicates an error condition where a specific
// semantic version was expected, but another version was present.
// This type is defined as an array containing two semantic version
// elements. The first element is the expected version and the second
// element is the actual version.
type MismatchingSemVerError [2]model.SemVer

// Error returns a string representation of `msve` error instance. This
// method causes *MismatchingSemVerError to implement error interface.
func (msve *MismatchingSemVerError) Error() string {
	expected := (*msve)[0]
	actual := (*msve)[1]
	return fmt.Sprintf(
		"expected v%s, but got v%s", expected.String(), actual.String(),
	)
}
