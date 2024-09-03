// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package settingsrp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/sch1v1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/core/log"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/usecase/appuc"
)

// Fetch queries the mutable settings from the settings repository,
// deserializes them, merges them into a clone of the baseConfs
// representing the configuration file and environment variables state,
// and returns the fresh configuration instance as an appuc.Builder
// interface in addition to its visible settings (as an instance of the
// version-independent model.VisibleSettings struct).
//
// The settings boundary values are also returned as `minb` and
// `maxb` instances (of the version-independent model.Settings
// struct), taken from the base settings. In order to sync these
// boundary values from the base settings to the database (so they
// can be queried by other components from the database), it is
// required to perform a migration or use the Update method of the
// SettingsTxQueryer interface instead. In other words, Fetch
// neither updates the database nor verifies it beyound the version
// of the persisted configuration settings.
// If the database settings were out of the acceptable range of
// values, they will take the nearest (minimum or maximum) boundary
// value and that adjustment will be logged as a warning.
func Fetch(
	ctx context.Context, c *postgres.Conn, baseConfs *cfg2.Config,
) (
	builder appuc.Builder,
	vs *model.VisibleSettings,
	minb, maxb *model.Settings,
	err error,
) {
	b, err := sch1v1.LoadSettings(ctx, c)
	if err != nil {
		err = fmt.Errorf("sch1v1.LoadSettings: %w", err)
		return nil, nil, nil, nil, err
	}
	var ser cfg2.Serializable
	err = json.Unmarshal(b, &ser)
	if err != nil {
		err = fmt.Errorf("deserializing json: %w", err)
		return nil, nil, nil, nil, err
	}
	confs := baseConfs.Clone()
	err = confs.Mutate(ser)
	var boundsErr settings.BoundsError
	switch {
	case errors.As(err, &boundsErr):
		log.Warn(
			ctx,
			"settings read from database are out of range",
			log.Err("err", err),
		)
	case err != nil:
		err = fmt.Errorf("confs.Mutate(%#v): %w", ser, err)
		return nil, nil, nil, nil, err
	}
	v := confs.Visible()
	vs = &model.VisibleSettings{
		ImmutableSettings: &model.ImmutableSettings{
			Logger: v.Immutable.Logger,
		},
	}
	if doo := v.Cars.DelayOfOPM; doo != nil {
		t := time.Duration(*doo)
		vs.ParkingMethod.Delay = &t
	}
	lb, ub := confs.Bounds()
	minb = adapterToModelSettings(lb)
	maxb = adapterToModelSettings(ub)
	return confs, vs, minb, maxb, nil
}

func adapterToModelSettings(s *cfg2.Serializable) *model.Settings {
	ms := &model.Settings{}
	if doo := s.Settings.Visible.Cars.DelayOfOPM; doo != nil {
		t := time.Duration(*doo)
		ms.VisibleSettings.ParkingMethod.Delay = &t
	}
	return ms
}

// Update converts the version-independent mutable model.Settings
// instance into a version-dependent serializable settings instance
// for the last supported version, serializes them as JSON, and
// then stores them in the settings repository. Given mutable settings
// are also used in order to update a clone of the baseConfs instance.
// Updated configuration settings will be returned as an instance of
// the appuc.Builder interface in addition to its visible settings
// (which are provided as an instance of the version-independent
// model.VisibleSettings struct).
//
// The settings boundary values are also returned as `minb` and
// `maxb` instances (of the version-independent model.Settings
// struct), taken from the base settings. The argument `s` settings
// must fall in this acceptable range of values, otherwise, an error
// will be returned and settings will be kept unchanged.
// When updating the database with new settings, the boundary values
// will be serialized and stored alongside them too.
func Update(
	ctx context.Context,
	tx *postgres.Tx,
	baseConfs *cfg2.Config,
	s *model.Settings,
) (
	builder appuc.Builder,
	vs *model.VisibleSettings,
	minb, maxb *model.Settings,
	err error,
) {
	ser := cfg2.Serializable{
		Version: cfg2.Version,
	}
	if d := s.VisibleSettings.ParkingMethod.Delay; d != nil {
		t := settings.Duration(*d)
		ser.Settings.Visible.Cars.DelayOfOPM = &t
	}
	confs := baseConfs.Clone()
	if err := confs.Mutate(ser); err != nil {
		// settings.BoundsError instances are handled here too
		err = fmt.Errorf("confs.Mutate(%#v): %w", ser, err)
		return nil, nil, nil, nil, err
	}
	b, err := json.Marshal(ser)
	if err != nil {
		err = fmt.Errorf("serializing json: %w", err)
		return nil, nil, nil, nil, err
	}
	lb, ub := confs.Bounds()
	lbb, err := json.Marshal(lb)
	if err != nil {
		err = fmt.Errorf("serializing lower bounds: %w", err)
		return nil, nil, nil, nil, err
	}
	ubb, err := json.Marshal(ub)
	if err != nil {
		err = fmt.Errorf("serializing upper bounds: %w", err)
		return nil, nil, nil, nil, err
	}
	sm1 := stlmig1.New(tx)
	err = sm1.PersistSettings(ctx, b, lbb, ubb)
	if err != nil {
		err = fmt.Errorf("persisting settings: %w", err)
		return nil, nil, nil, nil, err
	}
	v := confs.Visible()
	vs = &model.VisibleSettings{
		ImmutableSettings: &model.ImmutableSettings{
			Logger: v.Immutable.Logger,
		},
	}
	if doo := v.Cars.DelayOfOPM; doo != nil {
		t := time.Duration(*doo)
		vs.ParkingMethod.Delay = &t
	}
	minb = adapterToModelSettings(lb)
	maxb = adapterToModelSettings(ub)
	return confs, vs, minb, maxb, nil
}
