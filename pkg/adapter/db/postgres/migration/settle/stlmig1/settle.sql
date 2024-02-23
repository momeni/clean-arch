-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

SET search_path TO caweb1;

INSERT INTO cars (cid, name, lat, lon, parked, parking_mode)
SELECT cid, name, lat, lon, parked, parking_mode
    FROM mig1.cars;

INSERT INTO settings (component, config)
SELECT component, config
    FROM mig1.settings;
--  WHERE component!='caweb';
--  -- we do not exclude the caweb settings because having that row in
--  -- the database simplifies the SettingsPersister implementation
--  -- as it can always run an UPDATE instead of requiring to run
--  -- an INSERT or UPDATE or INSERT ON CONFLICT DO UPDATE based on
--  -- the database conditions.
