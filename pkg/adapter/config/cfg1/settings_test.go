// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cfg1_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg1"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/core/model"
)

func ExampleJSONSerialization() {
	s := &cfg1.Serializable{
		Version: model.SemVer{1, 4, 5},
	}
	opd := settings.Duration(2 * time.Minute)
	s.Settings.Visible.Cars.OldParkingDelay = &opd
	b, err := json.Marshal(s)
	fmt.Println(err)
	fmt.Println(string(b))
	// Output:
	// <nil>
	// {"version":"1.4.5","cars":{"old_parking_delay":"2m"}}
}

func ExampleJSONSerializationWithNilDuration() {
	s := &cfg1.Serializable{
		Version: model.SemVer{4, 1, 5},
	}
	b, err := json.Marshal(s)
	fmt.Println(err)
	fmt.Println(string(b))
	// Output:
	// <nil>
	// {"version":"4.1.5","cars":{"old_parking_delay":null}}
}
