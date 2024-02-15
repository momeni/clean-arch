-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

SET search_path TO mig1;

CREATE VIEW cars (cid, name, lat, lon, parked, parking_mode)
AS SELECT cid, name, lat, lon, parked, parking_mode
    FROM fdw1_1.cars;
