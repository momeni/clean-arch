// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package appuc

import (
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
)

// Builder interface represents the expectations from the application
// use case builders. All use cases which can be instantiated by a
// configuration struct have one NewX method here which takes database
// connection pool and their repository packages dependencies. All
// versions of configuration structs may be used by the migration use
// case, however, only the last supported version may be used by normal
// non-migration use cases. The last version of configuration struct
// must implement this interface, take repository packages and create
// use case objects based on its contained settings.
// When new settings are loaded from a database, they may override some
// of the configuration settings, hence, produce a new Builder instance.
type Builder interface {
	// NewAppUseCase creates a new application use case. This use case
	// needs a SettingsRepo in order to fetch or update mutable settings
	// from the database. It also needs to take all repository instances
	// which may be required by other use cases (e.g., a repo.Cars)
	// because it needs to pass them to the Builder instance again after
	// reloading or updating settings, changing the mutable settings in
	// the database and memory.
	//
	// When settings are updated and a new Builder instance is obtained,
	// it can be asked to create new use case objects. The fields of
	// the new application UseCase can be copied into the previous
	// UseCase instance, updating it in-place, while other use case
	// objects (e.g., carsuc.UseCase) can replace their old instance
	// as an opaque object. This replacement strategy requires the
	// resources packages to ask this application UseCase for the actual
	// use case objects, right before using them, so they can be fetched
	// or updated atomically as managed by the application UseCase.
	NewAppUseCase(
		p repo.Pool, s SettingsRepo, carsRepo repo.Cars,
	) (*UseCase, error)

	// NewCarsUseCase creates a new carsuc UseCase object having the
	// provided database connection pool and cars repository.
	NewCarsUseCase(p repo.Pool, r repo.Cars) (*carsuc.UseCase, error)
}
