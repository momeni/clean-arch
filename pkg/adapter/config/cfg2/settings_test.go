// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cfg2_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/core/model"
)

func ExampleJSONSerialization() {
	s := &cfg2.Serializable{
		Version: model.SemVer{1, 4, 5},
	}
	doo := settings.Duration(19 * time.Second)
	s.Settings.Visible.Cars.DelayOfOPM = &doo
	b, err := json.Marshal(s)
	fmt.Println(err)
	fmt.Println(string(b))
	// Output:
	// <nil>
	// {"version":"1.4.5","cars":{"delay_of_opm":"19s"}}
}

func ExampleJSONSerializationWithNilDuration() {
	s := &cfg2.Serializable{
		Version: model.SemVer{4, 1, 5},
	}
	b, err := json.Marshal(s)
	fmt.Println(err)
	fmt.Println(string(b))
	// Output:
	// <nil>
	// {"version":"4.1.5","cars":{"delay_of_opm":null}}
}
