// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cfg1

import (
	"errors"

	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/core/cerr"
	"github.com/momeni/clean-arch/pkg/core/model"
)

// Serializable embeds the Settings in addition to a Version field,
// so it can be serialized and stored in the database, while the Version
// field may be consulted during its deserialization in order to ensure
// that it belongs to the same configuration format version.
// The Serializable and the main Config struct are versioned together.
// The nested Immutable pointer must be nil because the Serializable
// is supposed to carry the mutable settings which are acceptable to be
// queried from the database and may be passed to the Mutate method.
type Serializable struct {
	// Version indicates the format version of this Serializable and
	// is equal to the Config struct version. Although its value is
	// known from the Serializable type, but we have to store it as a
	// field in order to find it out during the deserialization and
	// application phase (by the Mutate method).
	// Therefore, the embedded Settings struct is enough at runtime.
	Version model.SemVer `json:"version"`

	Settings
}

// Settings contains those settings which are mutable & invisible,
// that is, write-only settings. It also embeds the Visible struct
// so it effectively contains all kinds of settings. When fetching
// settings, the nested Immutable pointer can be set to nil in order to
// keep the mutable (visible or invisible) settings and when reporting
// settings, the embedded Visible struct can be reported alone (having
// a non-nil Immutable pointer) in order to exclude the invisible
// settings.
//
// Some fields, such as Logger, were defined as a pointer in the Config
// struct because it was desired to detect if they were or were not
// initialized during a migration operation, so they could be filled
// by the MergeConfig method later. They had to obtain a value anyways
// after a call to the ValidateAndNormalize method and so nil is not a
// meaningful value for them. Those fields must have non-pointer types
// in the Settings and Visible structs, so they take a value when read
// from the database for example. Even if a settings manipulation use
// case implementation wants to allow end-users to selectively configure
// settings, it is the responsibility of that implementation to replace
// such nil values with their old settings values and we can expect to
// set all fields of the Settings and Visible structs collectively.
// By the way, such a use case increases the risk of conflicts because
// an end-user decides to selectively update one setting because they
// think that other settings have some seen values, but they have been
// changed concurrently. So it is preferred to ask the frontend to send
// the complete set of settings (whether they are set by end-user or
// their older seen values are left unchanged) in order to justify a PUT
// instead of a POST request method. Of course, that decision relies on
// the details of each use case and cannot be fixed in this layer.
//
// Some fields, such as the old parking method delay, were defined as a
// pointer in the Config struct because they could be left uninitialized
// even after a call to the ValidateAndNormalize method. That is, nil
// is a meaningful value for them and asks the configuration instance
// not to pass their corresponding functional options to use cases.
// Those fields must have pointer types in the Settings and Visible
// structs, so they can be kept uninitialized even when stored in and
// read out from the database again. That is, even if a settings field
// has a non-nil value, but its corresponding field in the database
// has a nil value, it has to be overwritten by that nil because being
// uninitialized is a menaingful configuration decision which was taken
// and persisted in the database in that scenario.
type Settings struct {
	Visible
}

// Visible contains settings which are visible by end-users.
// These settings may be mutable or immutable. The immutable & visible
// settings are managed by the embedded Immutable struct. When it is
// desired to serialize and transmit settings to end-users, the
// Immutable pointer should be non-nil and its fields should be
// poppulated. However, when it is desired to fetch settings from
// end-users and deserialize them, the Immutable pointer should be set
// to nil in order to abandon them.
type Visible struct {
	// Cars represents the visible and mutable settings for the Cars
	// use cases.
	Cars struct {
		// OldParkingDelay indicates the old parking method delay.
		OldParkingDelay *settings.Duration `json:"old_parking_delay"`
	} `json:"cars"`
	*Immutable
}

// Immutable contains settings which are immutable (and can be
// configured only using the configuration file or environment variables
// alone), but are visible by end-users (settings must be at least
// visible or mutable, otherwise, they may not be called a setting).
type Immutable struct {
	// Logger reports if server-side REST API logging is enabled.
	Logger bool `json:"logger"`
}

// Mutate updates this Config instance using the given Serializable
// instance which provides the mutable settings values.
// The given Serializable instance may contain mutable & invisible
// settings (write-only) and mutable & visible settings (read-write),
// but it may not contain the immutable settings (i.e., the Immutable
// pointer must be nil). The provided Serializable instance is not
// updated itself, hence, a non-pointer variable is suitable.
func (c *Config) Mutate(s Serializable) error {
	if s.Settings.Visible.Immutable != nil {
		return errors.New("immutable settings must not be set")
	}
	if v1 := c.Version(); v1 != s.Version {
		return &cerr.MismatchingSemVerError{v1, s.Version}
	}
	settings.OverwriteUnconditionally(
		&c.Usecases.Cars.OldParkingDelay,
		s.Settings.Visible.Cars.OldParkingDelay,
	)
	return nil
}

// Serializable creates and returns an instance of *Serializable
// in order to report the mutable settings, based on this Config
// instance. The Immutable pointer will be nil in the returned object.
func (c *Config) Serializable() *Serializable {
	s := &Serializable{
		Version: c.Version(),
		Settings: Settings{
			Visible: Visible{
				Immutable: nil,
			},
		},
	}
	settings.OverwriteUnconditionally(
		&s.Settings.Visible.Cars.OldParkingDelay,
		c.Usecases.Cars.OldParkingDelay,
	)
	return s
}

// Visible creates and fills an instance of Visible struct with the
// mutable and immutable settings which can be queried by end-users.
// That is, the Immutable pointer will be non-nil in the returned
// object. Despite the Mutate and Serializable methods, the Visible
// method is not included in the pkg/adapter/config/settings.Config
// generic interface because it is only useful in the adapters layer
// where a repository package may query the visible settings after
// updating a Config instance. However, it is not required in the
// migration use cases as they deal with mutable settings which are
// exposed by the Serializable method.
func (c *Config) Visible() *Visible {
	v := &Visible{
		Immutable: &Immutable{
			// The panic on nil-dereference of c.Gin.Logger is fine
			// because after a call to the ValidateAndNormalize method,
			// Logger must be non-nil (in absence of programming errors)
			// and this is the reason that Logger in Immutable struct is
			// not defined as a pointer itself (while OldParkingDelay
			// field is defined as a pointer).
			Logger: *c.Gin.Logger,
		},
	}
	settings.OverwriteUnconditionally(
		&v.Cars.OldParkingDelay, c.Usecases.Cars.OldParkingDelay,
	)
	return v
}
