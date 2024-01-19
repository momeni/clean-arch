// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

// Coordinate represents a geographical location with a latitude and
// longitude. This struct is included in the Car struct in order to
// demonstrate how a struct may be embedded while mapping them to a
// database table.
type Coordinate struct {
	Lat, Lon float64 // latitude and longitude of the geo-location
}
