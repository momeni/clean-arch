-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

SET search_path TO caweb1;

INSERT INTO cars (cid, name, lat, lon, parked, parking_mode)
SELECT cid, name, lat, lon, parked, parking_mode
    FROM mig1.cars;
