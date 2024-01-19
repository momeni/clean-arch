-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

CREATE TABLE cars (
    cid uuid NOT NULL,
    name text,
    lat numeric,
    lon numeric,
    parked boolean
);

ALTER TABLE ONLY cars
    ADD CONSTRAINT cars_pkey PRIMARY KEY (cid);
