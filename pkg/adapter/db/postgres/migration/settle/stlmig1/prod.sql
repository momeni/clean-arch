-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

-- The initial data rows which are required in a production environment
-- should be added here.

SET search_path TO caweb1;

-- The version (for each configuration format major version) must match
-- with the latest supported minor version.
INSERT INTO settings (component, config, min_bounds, max_bounds)
VALUES (
    'caweb',
    '{"version":"2.1.0","cars":{"delay_of_opm":"12s"}}'::json,
    '{"version":"2.1.0","cars":{"delay_of_opm":"1s"}}'::json,
    '{"version":"2.1.0","cars":{"delay_of_opm":"7h"}}'::json
);
