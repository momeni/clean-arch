// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package appuc

import (
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
)

// Settings returns a copy of visible settings which are currently in
// effect. The effective settings and use case objects which are built
// based on them (and other invisible settings) may be updated
// atomically, while they are exposed by a series of getter methods. At
// least one of Reload or UpdateSettings methods must be called before
// this (and other use case objects getter methods) may be called.
func (app *UseCase) Settings() model.VisibleSettings {
	app.rwlock.RLock()
	defer app.rwlock.RUnlock()
	return *app.settings
}

// updateAll atomically updates the visible settings and all other use
// case objects which are built based on thise (visible and invisible)
// settings. This method minimizes the scope which needs to take a
// writing lock (after instantiating all relevant use case objects).
func (app *UseCase) updateAll(
	vs *model.VisibleSettings,
	carsUseCase *carsuc.UseCase,
) {
	app.rwlock.Lock()
	defer app.rwlock.Unlock()
	app.settings = vs
	app.carsUseCase = carsUseCase
}

// CarsUseCase returns the currently effective cars use case object. At
// least one of Reload or UpdateSettings methods must be called before
// this (and other use case objects getter methods) may be called.
func (app *UseCase) CarsUseCase() *carsuc.UseCase {
	app.rwlock.RLock()
	defer app.rwlock.RUnlock()
	return app.carsUseCase
}
