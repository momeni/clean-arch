// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package appuc

import (
	"context"
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
)

// UpdateSettings updates the mutable settings in the database, with
// help of a settings repository instance, according to the given `s`
// settings. This leads to preparation of a fresh Builder instance
// which can be used for creation of fresh use case objects, including
// the application use case itself. Thereafter, new visible settings
// and use case objects will be changed atomically to their fresh
// values. Despite other use case objects which can be replaced by new
// instances, whenever the application use case had some dependency on
// settings, we have to take those relevant fields from the new appuc
// UseCase instance and update the `app` instance fields in-place.
// The reason is that other use case objects are fetched by resources
// packages before each request (using a synchronized getter of this
// application use case instance), however, there is no method
// to replace the `app` instance itself.
//
// UpdateSettings and Reload methods are synchronized using a mutex
// so only one long-running attempt for querying/updating the mutable
// settings may exist, while other goroutines may fetch the old settings
// and use case objects without any blocking. When the operation could
// complete successfully and new use case objects were created, a second
// read-write lock will be used in order to pause other goroutines and
// switch all use case objects to new instances. The order of these
// locks ensures a deadlock-free implementation.
//
// The returned `vs` vissible settings and `minb` and `maxb` which are
// minimum/maximum boundary settings values are pointers to the shared
// structs. If caller needs to modify them, those structs must be deeply
// cloned beforehand.
func (app *UseCase) UpdateSettings(
	ctx context.Context, s *model.Settings,
) (vs *model.VisibleSettings, minb, maxb *model.Settings, err error) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	var managed managedUseCases
	err = app.pool.Conn(
		ctx, func(ctx context.Context, c repo.Conn) error {
			return c.Tx(
				ctx, func(ctx context.Context, tx repo.Tx) error {
					q := app.settingsRepo.Tx(tx)
					var b Builder
					b, vs, minb, maxb, err = q.Update(ctx, s)
					if err != nil {
						return fmt.Errorf("database update: %w", err)
					}
					managed, err = app.newManagedUseCases(b)
					if err != nil {
						return fmt.Errorf("creating use cases: %w", err)
					}
					return nil
				},
			)
		},
	)
	if err != nil {
		err = fmt.Errorf("delegating update to settings repo: %w", err)
		return nil, nil, nil, err
	}
	app.updateAll(vs, minb, maxb, managed)
	return vs, minb, maxb, nil
}

// Reload queries the settings repository in order to fetch the current
// effective mutable settings. Those settings will override the base
// settings which were read from a configuration file (and possibly
// overridden by environment variables) in order to create a fresh
// Builder instance. Thereafter, that Builder instance will be used for
// creation of fresh use case objects, including the application use
// case itself. Ultimately, new visible settings and use case objects
// will be changed atomically to their fresh values. Despite other
// use case objects which can be replaced by new instances, whenever
// the application use case had some dependency on settings, we have to
// take those relevant fields from the new appuc UseCase instance and
// update the `app` instance fields in-place.
// The reason is that other use case objects are fetched by resources
// packages before each request (using a synchronized getter of this
// application use case instance), however, there is no method
// to replace the `app` instance itself.
//
// UpdateSettings and Reload methods are synchronized using a mutex
// so only one long-running attempt for querying/updating the mutable
// settings may exist, while other goroutines may fetch the old settings
// and use case objects without any blocking. When the operation could
// complete successfully and new use case objects were created, a second
// read-write lock will be used in order to pause other goroutines and
// switch all use case objects to new instances. The order of these
// locks ensures a deadlock-free implementation.
func (app *UseCase) Reload(ctx context.Context) error {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	var (
		b          Builder
		vs         *model.VisibleSettings
		minb, maxb *model.Settings
		err        error
	)
	err = app.pool.Conn(
		ctx, func(ctx context.Context, c repo.Conn) error {
			q := app.settingsRepo.Conn(c)
			b, vs, minb, maxb, err = q.Fetch(ctx)
			return err
		},
	)
	if err != nil {
		return fmt.Errorf("reloading by settings repo: %w", err)
	}
	managed, err := app.newManagedUseCases(b)
	if err != nil {
		return fmt.Errorf("creating use cases: %w", err)
	}
	app.updateAll(vs, minb, maxb, managed)
	return nil
}

// newManagedUseCases creates all relevant use case objects using the
// given Builder instance and wraps their pointers by a managedUseCases
// struct, so they may be passed to the updateAll method later.
// The updateAll is not called directly because after a successful
// instantiation of all use case objects, we may need to wait for a
// database transaction to commit yet.
func (app *UseCase) newManagedUseCases(
	b Builder,
) (managedUseCases, error) {
	var nilm managedUseCases
	carsUseCase, err := b.NewCarsUseCase(app.pool, app.carsRepo)
	if err != nil {
		return nilm, fmt.Errorf("creating cars use case: %w", err)
	}
	return managedUseCases{
		carsUseCase: carsUseCase,
	}, nil
}
