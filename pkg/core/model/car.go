// Copyright (c) 2023 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package model defines the inner most layer of the Clean Architecture
// containing the business-level models, also called entities or domain.
// This layer may not depend on outter layers, while all other layers
// may depend on it.
// By the way, it is acceptable to annotate structs in this package with
// multiple frameworks dependent tags (e.g., as required by ORM
// libraries) since adding more tags does not complicate definition of
// a struct, but can prevent unnecessary structs duplication.
package model

// Car models a car which may be persisted in a database.
// This model does not contain an ID in order to demonstrate that how
// a model which has no tags and its fields do not match with the
// expected table may be managed by the adapter layer.
// For the corresponding struct which fixes these issues and stores the
// resulting struct in the database, see the unexported gCar struct
// in the pkg/adapter/db/postgres/carsrp/query.go file.
type Car struct {
	Name       string     // name of the car
	Coordinate Coordinate // current location of car
	Parked     bool       // a flag to indicate if car is parked/moving
}
