-- Copyright (c) 2023-2024 Behnam Momeni
-- This Source Code Form is subject to the terms of the Mozilla Public
-- License, v. 2.0. If a copy of the MPL was not distributed with this
-- file, You can obtain one at https://mozilla.org/MPL/2.0/.

SET search_path TO caweb1;

CREATE TABLE cars (
    cid uuid NOT NULL,
    name text NOT NULL,
    lat numeric NOT NULL,
    lon numeric NOT NULL,
    parked boolean NOT NULL,
    parking_mode text
);

ALTER TABLE ONLY cars
ADD CONSTRAINT cars_pkey PRIMARY KEY (cid);

CREATE TABLE settings (
    -- an enum type instead of text may be helpful here too
    component text NOT NULL,
    -- json type is preferred over jsonb because we expect no reparsing
    -- and searching on this field; indeed, a text type might work with
    -- similar performance, but with less clarity about the contents of
    -- this column and no validation (and of course, we can change it
    -- to jsonb by a minor version update whenever searching in the
    -- database became helpful).
    config json NOT NULL
);

ALTER TABLE ONLY settings
ADD CONSTRAINT settings_pkey PRIMARY KEY (component);
