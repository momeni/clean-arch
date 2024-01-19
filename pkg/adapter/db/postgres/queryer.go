// Copyright (c) 2023-2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package postgres

import (
	"context"

	"github.com/momeni/clean-arch/pkg/core/repo"
	"gorm.io/gorm"
)

// Queryer interface includes methods for running SQL statements.
// There are two main types of statements. One category may affect
// multiple rows, but do not return a result set, like DDL commands
// or an UPDATE without a RETURNING clause. These statements are
// executed with the Exec method. Another category of statements
// (which may or may not modify the database contents) provide a result
// set, like SELECT or an UPDATE with a RETURNING clause. These queries
// may be executed with the Query method.
// This interface is embedded by both of Conn and Tx since they may be
// used for execution of commands, of course, with distinct isolation
// levels.
//
// Queryer is used as a generic type constraint and may be satisfied
// by *Conn or *Tx. The common interface between these two types
// includes the repo.Queryer in addition to the GORM(ctx) method which
// allows its clients to access the embedded *gorm.DB instance from
// generic functions.
type Queryer interface {
	*Conn | *Tx
	repo.Queryer

	// GORM returns the embedded *gorm.DB instance, configuring it
	// to operate on the given ctx context (in a gorm.Session).
	GORM(ctx context.Context) *gorm.DB
}
