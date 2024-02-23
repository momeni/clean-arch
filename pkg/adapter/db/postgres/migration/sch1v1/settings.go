// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sch1v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/momeni/clean-arch/pkg/core/repo"
)

// LoadSettings loads the serialized mutable settings from the database
// using the given `c` connection, assuming that the database schema
// version is equal to v1.1 as specified by the Major and Minor consts.
func LoadSettings(ctx context.Context, c repo.Conn) ([]byte, error) {
	rs, err := c.Query(
		ctx, "SELECT config FROM settings WHERE component='caweb'",
	)
	if err != nil {
		return nil, fmt.Errorf("querying settings table: %w", err)
	}
	defer rs.Close()
	var cfg []byte
	for rs.Next() {
		if cfg != nil {
			return nil, errors.New("more than one caweb settings rows")
		}
		if err := rs.Scan(&cfg); err != nil {
			return nil, fmt.Errorf("scanning config column: %w", err)
		}
	}
	if err := rs.Err(); err != nil {
		return nil, fmt.Errorf("closing result set: %w", err)
	}
	if cfg == nil {
		return nil, errors.New("missing caweb settings row")
	}
	return cfg, nil
}
