// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cfg1_test

import (
	"fmt"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg1"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/config/vers"
	"github.com/momeni/clean-arch/pkg/core/model"
	"gopkg.in/yaml.v3"
)

// This conversion ensures that *cfg1.Config implements the generic
// settings.Config interface. Such tests should be used when the actual
// implementation does not take a type as its expected interface and so
// it may be converted to a version-independent interface without first
// being converted to a version-specific interface (and so causing a
// compilation error in case of the mismatched types instead of getting
// some runtime error).
var _ settings.Config[*cfg1.Config, cfg1.Serializable] = (*cfg1.Config)(nil)

func ExampleMarshalYAML() {
	d, l, r := settings.Duration(time.Hour), true, true
	minb := settings.Duration(time.Second)
	maxb := settings.Duration(5 * time.Hour)
	c := &cfg1.Config{
		Database: cfg1.Database{
			Host:    "127.0.0.1",
			Port:    1234,
			Name:    "caweb1_0_0",
			PassDir: "/var/lib/caweb/db/caweb1_0_0",
		},
		Gin: cfg1.Gin{
			Logger:   &l,
			Recovery: &r,
		},
		Usecases: cfg1.Usecases{
			Cars: cfg1.Cars{
				OldParkingDelay:    &d,
				MinOldParkingDelay: &minb,
				MaxOldParkingDelay: &maxb,
			},
		},
		Vers: vers.Config{
			Versions: vers.Versions{
				Database: model.SemVer{1, 4, 5},
				Config:   model.SemVer{5, 4, 1},
			},
		},
	}
	b, err := yaml.Marshal(c)
	fmt.Println(err)
	fmt.Println(string(b))
	// Output:
	// <nil>
	// database:
	//     host: 127.0.0.1
	//     port: 1234
	//     name: caweb1_0_0
	//     pass-dir: /var/lib/caweb/db/caweb1_0_0
	// gin:
	//     logger: true
	//     recovery: true
	// usecases:
	//     cars:
	//         old-parking-method-delay: 1h
	//         old-parking-method-delay-minimum: 1s
	//         old-parking-method-delay-maximum: 5h
	// versions:
	//     database: 1.4.5
	//     config: 5.4.1
}
