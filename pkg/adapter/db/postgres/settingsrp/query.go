// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package settingsrp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/momeni/clean-arch/pkg/adapter/config/cfg2"
	"github.com/momeni/clean-arch/pkg/adapter/config/settings"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/sch1v1"
	"github.com/momeni/clean-arch/pkg/adapter/db/postgres/migration/settle/stlmig1"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/usecase/appuc"
)

// Fetch queries the mutable settings from the settings repository,
// deserializes them, merges them into a clone of the baseConfs
// representing the configuration file and environment variables state,
// and returns the fresh configuration instance as an appuc.Builder
// interface in addition to its visible settings (as an instance of the
// version-independent model.VisibleSettings struct).
func Fetch(
	ctx context.Context, c *postgres.Conn, baseConfs *cfg2.Config,
) (appuc.Builder, *model.VisibleSettings, error) {
	b, err := sch1v1.LoadSettings(ctx, c)
	if err != nil {
		return nil, nil, fmt.Errorf("sch1v1.LoadSettings: %w", err)
	}
	var ser cfg2.Serializable
	err = json.Unmarshal(b, &ser)
	if err != nil {
		return nil, nil, fmt.Errorf("deserializing json: %w", err)
	}
	confs := baseConfs.Clone()
	err = confs.Mutate(ser)
	if err != nil {
		return nil, nil, fmt.Errorf("confs.Mutate(%#v): %w", ser, err)
	}
	v := confs.Visible()
	vs := &model.VisibleSettings{
		ImmutableSettings: &model.ImmutableSettings{
			Logger: v.Immutable.Logger,
		},
	}
	if doo := v.Cars.DelayOfOPM; doo != nil {
		t := time.Duration(*doo)
		vs.ParkingMethod.Delay = &t
	}
	return confs, vs, nil
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
func Update(
	ctx context.Context,
	tx *postgres.Tx,
	baseConfs *cfg2.Config,
	s *model.Settings,
) (appuc.Builder, *model.VisibleSettings, error) {
	ser := cfg2.Serializable{
		Version: cfg2.Version,
	}
	if d := s.VisibleSettings.ParkingMethod.Delay; d != nil {
		t := settings.Duration(*d)
		ser.Settings.Visible.Cars.DelayOfOPM = &t
	}
	confs := baseConfs.Clone()
	if err := confs.Mutate(ser); err != nil {
		return nil, nil, fmt.Errorf("confs.Mutate(%#v): %w", ser, err)
	}
	b, err := json.Marshal(ser)
	if err != nil {
		return nil, nil, fmt.Errorf("serializing json: %w", err)
	}
	sm1 := stlmig1.New(tx)
	err = sm1.PersistSettings(ctx, b)
	if err != nil {
		return nil, nil, fmt.Errorf("persisting settings: %w", err)
	}
	v := confs.Visible()
	vs := &model.VisibleSettings{
		ImmutableSettings: &model.ImmutableSettings{
			Logger: v.Immutable.Logger,
		},
	}
	if doo := v.Cars.DelayOfOPM; doo != nil {
		t := time.Duration(*doo)
		vs.ParkingMethod.Delay = &t
	}
	return confs, vs, nil
}
