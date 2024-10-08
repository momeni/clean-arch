-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

SET search_path TO mig1;

CREATE VIEW cars (cid, name, lat, lon, parked, parking_mode)
AS SELECT
        cid, name, lat, lon, parked,
        CASE
            WHEN parked = true THEN 'old'
            ELSE NULL
        END
    FROM fdw1_0.cars;

CREATE VIEW settings (component, config, min_bounds, max_bounds)
AS SELECT component, config,
        json_object('version': config->>'version'),
        json_object('version': config->>'version')
    FROM fdw1_0.settings;
