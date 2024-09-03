// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cfg2

import (
	"errors"
	"fmt"

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
//
// Serializable also can represent the minimum and maximum boundary
// values for settings. Because all mutable and immutable settings can
// have boundary values potentially, all fields may have a value in
// this use case. The version of the main settings and its boundary
// values (i.e., three instances of this struct) must be the same.
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
// meaningful value for them. However, they must have a pointer type
// yet because Settings instances can represent the boundary values too.
// All fields which may need a minimum or maximum boundary value must be
// defined as a pointer so they can be left uninitialized when no such
// limit is desired. Using a non-pointer type means that a minimum and a
// maximum value must be provided for that setting (which is not
// possible for data types which are not ordered).
// Note that a nil value may not be used to communicate a scenario that
// user does not want to modify a setting. If a use case implementation
// wants to allow end-users to selectively configure settings, it is
// the responsibility of that implementation to replace such nil values
// with their old settings values and we can expect to set all fields
// of the Settings and Visible structs collectively. This is essential
// so caller can explicitly deinitialize a setting and enable its
// use case layer default value. If a use case requires to distinguish
// between deinitializing and not configuring a setting at all, an extra
// boolean setting may be sent such as `setting` and `setting_set` which
// its false value indicates that no setting is sent and its true value
// indicates that a value (which can be nil too) is sent.
// By the way, a use case which does not set some setting without having
// a UX meaning for it (such as a switch button for removing some
// restriction), increases the risk of conflicts because
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
//
// Therefore, all fields must have pointer types, although some of them
// must be always non-nil (or they will poppulate an invalid value).
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
		// DelayOfOPM indicates the old parking method delay.
		//
		// When used as a minimum or maximum boundary value, it will
		// indicate the smallest/largest acceptable delay, or no such
		// restriction if set to nil.
		DelayOfOPM *settings.Duration `json:"delay_of_opm"`
	} `json:"cars"`
	*Immutable
}

// Immutable contains settings which are immutable (and can be
// configured only using the configuration file or environment variables
// alone), but are visible by end-users (settings must be at least
// visible or mutable, otherwise, they may not be called a setting).
type Immutable struct {
	// Logger reports if server-side REST API logging is enabled.
	//
	// This field must always have a non-nil value when it represents
	// the setting value and must always be nil when it represents the
	// boundary values.
	Logger *bool `json:"logger"`
}

// OutOfBoundsSettingsError has the same structure as the Serializable
// struct and its embedded structs with three differences:
//  1. Fields are listed directly in OutOfBoundsSettingsError struct
//     instead of being categorized based on their immutability and
//     visibility status, because all of those settings may have an
//     out of range error and all such errors should be reported
//     together,
//  2. Only that subset of fields is included which may observe a
//     *settings.OutOfRangeError[T] error, because if other errors could
//     happen, they would be reported by higher priority and they would
//     obstruct the mutation request, while an OutOfBoundsSettingsError
//     is returned by the Mutate method only when the caller is free
//     to decide if error should be fatal or treated as a warning,
//  3. The type of all included fields is *settings.OutOfRangeError[T]
//     for different T types, where T is the actual type of that field
//     from the Serializable struct.
type OutOfBoundsSettingsError struct {
	// Cars contains errors related to the cars use cases.
	Cars struct {
		// DelayOfOPM indicates the range violation error (if any)
		// with regards to the old parking method delay.
		DelayOfOPM *settings.OutOfRangeError[settings.Duration]
	}
}

// Error implements error interface and encodes whole of this
// OutOfBoundsSettingsError instance as an error string.
func (e *OutOfBoundsSettingsError) Error() string {
	return fmt.Sprintf("cfg2.Config settings are out of bounds: %#v", e)
}

// IsBoundsError implements the settings.BoundsError interface and
// so marks the *OutOfBoundsSettingsError as a boundary values violation
// error.
func (e *OutOfBoundsSettingsError) IsBoundsError() {
}

// Mutate updates this Config instance using the given Serializable
// instance which provides the mutable settings values.
// The given Serializable instance may contain mutable & invisible
// settings (write-only) and mutable & visible settings (read-write),
// but it may not contain the immutable settings (i.e., the Immutable
// pointer must be nil). The provided Serializable instance is not
// updated itself, hence, a non-pointer variable is suitable.
//
// If provided values do not respect the expected boundary values, an
// error will be returned, indicating that which settings were out of
// bound, however, this type of error does not prevent this Config
// instance to be updated. When a minimum/maximum boundary value is
// crossed over, that boundary value itself will be used as the new
// value of that setting. In this scenario, returned error will have
// the *OutOfBoundsSettingsError type.
func (c *Config) Mutate(s Serializable) error {
	if s.Settings.Visible.Immutable != nil {
		return errors.New("immutable settings must not be set")
	}
	if v1 := c.Version(); v1 != s.Version {
		return &cerr.MismatchingSemVerError{v1, s.Version}
	}
	settings.OverwriteUnconditionally(
		&c.Usecases.Cars.DelayOfOPM, s.Settings.Visible.Cars.DelayOfOPM,
	)
	boundsErr, hasBoundsErr := &OutOfBoundsSettingsError{}, false
	if err := settings.VerifyRange(
		&c.Usecases.Cars.DelayOfOPM,
		c.Usecases.Cars.MinDelayOfOPM,
		c.Usecases.Cars.MaxDelayOfOPM,
	); err != nil {
		boundsErr.Cars.DelayOfOPM = err
		hasBoundsErr = true
	}
	if hasBoundsErr {
		return boundsErr
	}
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
		&s.Settings.Visible.Cars.DelayOfOPM, c.Usecases.Cars.DelayOfOPM,
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
	// The panic on nil-dereference of c.Gin.Logger is fine because
	// after a call to the ValidateAndNormalize method, Logger must be
	// non-nil (in absence of programming errors).
	l := *c.Gin.Logger
	v := &Visible{
		Immutable: &Immutable{
			Logger: &l,
		},
	}
	settings.OverwriteUnconditionally(
		&v.Cars.DelayOfOPM, c.Usecases.Cars.DelayOfOPM,
	)
	return v
}

// Bounds creates and returns two instances of *Serializable in order to
// report the minimum and maximum boundary values for those settings
// which their lower/upper limits should be restricted.
// The boundary values may be reported for both of the mutable and
// immutable settings (as they have an informational purpose).
// All boundary values are obtained from this Config instance.
func (c *Config) Bounds() (minb, maxb *Serializable) {
	minb = &Serializable{
		Version: c.Version(),
		Settings: Settings{
			Visible: Visible{
				Immutable: &Immutable{},
			},
		},
	}
	settings.OverwriteUnconditionally(
		&minb.Settings.Visible.Cars.DelayOfOPM,
		c.Usecases.Cars.MinDelayOfOPM,
	)
	maxb = &Serializable{
		Version: c.Version(),
		Settings: Settings{
			Visible: Visible{
				Immutable: &Immutable{},
			},
		},
	}
	settings.OverwriteUnconditionally(
		&maxb.Settings.Visible.Cars.DelayOfOPM,
		c.Usecases.Cars.MaxDelayOfOPM,
	)
	return minb, maxb
}
