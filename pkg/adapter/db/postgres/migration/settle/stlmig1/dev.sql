-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

-- The initial data rows which are required in a development environment
-- should be added here.

SET search_path TO caweb1;

INSERT INTO cars(cid, name, lat, lon, parked, parking_mode)
VALUES (
        'e4f6b292-5dfe-4877-9cd2-7575d95825a8',
        'Bugatti',
        26.239947,
        55.147466,
        true,
        'old'
    ), (
        '024541b0-7aa7-4468-97f7-66b13ed25e04',
        'Maserati',
        25.878869,
        55.021838,
        true,
        'new'
    ), (
        '7c7d505d-2181-4352-90a8-c1426ed19159',
        'Nissan',
        25.880152,
        55.023427,
        false,
        NULL
    );

-- The version (for each configuration format major version) must match
-- with the latest supported minor version.
INSERT INTO settings (component, config, min_bounds, max_bounds)
VALUES (
    'caweb',
    '{"version":"2.1.0","cars":{"delay_of_opm":"2s"}}'::json,
    '{"version":"2.1.0","cars":{"delay_of_opm":"1s"}}'::json,
    '{"version":"2.1.0","cars":{"delay_of_opm":"7h"}}'::json
);
