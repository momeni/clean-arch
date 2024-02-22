// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package model

import "time"

// Settings contains those settings which are mutable & invisible,
// that is, write-only settings. It also embeds the VisibleSettings
// struct, so it effectively contains all kinds of settings.
// When fetching settings, the nested ImmutableSettings pointer can be
// set to nil in order to keep the mutable (visible or invisible)
// settings and when reporting settings, the embedded VisibleSettings
// struct can be reported alone (having a non-nil ImmutableSettings
// pointer) in order to exclude the invisible settings.
//
// This model layer struct is required (in addition to its version
// dependent adapters layer counterparts) because settings should be
// reported to and taken from end-users as required from the use cases
// layer. A repository package is responsible to manage conversion
// between these structs (only supporting the latest configuration
// version at any time).
type Settings struct {
	VisibleSettings
}

// VisibleSettings contains settings which are visible by end-users.
// These settings may be mutable or immutable. The immutable & visible
// settings are managed by the embedded ImmutableSettings struct.
// When it is desired to serialize and transmit settings to end-users,
// the ImmutableSettings pointer should be non-nil and its fields should
// be poppulated. However, when it is desired to fetch settings from
// end-users and deserialize them, the ImmutableSettings pointer should
// be set to nil in order to abandon them.
//
// This model layer struct is required (in addition to its version
// dependent adapters layer counterparts) because settings should be
// reported to and taken from end-users as required from the use cases
// layer. A repository package is responsible to manage conversion
// between these structs (only supporting the latest configuration
// version at any time).
type VisibleSettings struct {
	// ParkingMethod contains the old parking method related settings.
	ParkingMethod ParkingMethodSettings `json:"parking_method"`

	*ImmutableSettings
}

// ParkingMethodSettings represents the old parking method related
// settings. These settings are considered both visible and mutable.
type ParkingMethodSettings struct {
	// Delay represents the old parking method delay.
	Delay *time.Duration `json:"delay"`
}

// ImmutableSettings contains settings which are immutable (and can be
// configured only using the configuration file or environment variables
// alone), but are visible by end-users (settings must be at least
// visible or mutable, otherwise, they may not be called a setting).
//
// This model layer struct is required (in addition to its version
// dependent adapters layer counterparts) because settings should be
// reported to and taken from end-users as required from the use cases
// layer. A repository package is responsible to manage conversion
// between these structs (only supporting the latest configuration
// version at any time).
type ImmutableSettings struct {
	// Logger reports if server-side REST API logging is enabled.
	Logger bool `json:"logger"`
}
