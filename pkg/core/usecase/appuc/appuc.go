// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package appuc contains the application UseCase which supports the
// settings fetching and updating requests, allows the application to be
// reloaded based on the mutable settings which are stored in the
// database, and maintains and provides visible settings and use case
// objects (with atomic replacement support) so they may be used by
// the resources packages.
package appuc

import (
	"fmt"
	"sync"

	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"github.com/momeni/clean-arch/pkg/core/usecase/carsuc"
)

// UseCase represents an application use case. It holds a database
// connection pool, settings repository instance, and all repository
// instances which are required by other supported use cases.
// Therefore, it can pass these repository instances to a use case
// builder object (which is realized by the effective Config instance)
// in order to create supported use case objects (during a Reload or
// UpdateSettings operation).
type UseCase struct {
	pool         repo.Pool
	settingsRepo SettingsRepo
	carsRepo     repo.Cars

	// mutex is used by UpdateSettings and Reload methods so only one
	// go routine can try to update/fetch settings from the database
	// at any time. Even though such update/query attempts may proceed
	// concurrently as while as the UPDATE/SELECT queries are concerned,
	// but there is a risk that a go routine obtaining older data fail
	// to call updateAll sooner, hence, the older data may last longer.
	// The mutex resolves such concurrency issues without blocking other
	// use cases (which should be fetched using the following rwlock).
	mutex sync.Mutex

	// rwlock is locked for writing by updateAll whenever the new state
	// including the visible settings and use case objects are prepared
	// and should be published atomically, while it is locked by all
	// getter methods for reading in order to access the published state
	// (i.e., visible settings and use case objects).
	rwlock sync.RWMutex

	settings    *model.VisibleSettings // cached visible settings
	carsUseCase *carsuc.UseCase
}

// New instantiates an application use case object. The Reload method
// of this object should be called at least once, so it can create
// other supported use case objects, before their corresponding getter
// methods are invoked (otherwise, they may return nil).
func New(
	p repo.Pool, s SettingsRepo, carsRepo repo.Cars, opts ...Option,
) (*UseCase, error) {
	uc := &UseCase{
		pool:         p,
		settingsRepo: s,
		carsRepo:     carsRepo,
	}
	for _, opt := range opts {
		if err := opt(uc); err != nil {
			return nil, fmt.Errorf("invalid option: %w", err)
		}
	}
	return uc, nil
}
