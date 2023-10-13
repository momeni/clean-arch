
CREATE TABLE cars (
    cid uuid NOT NULL,
    name text,
    lat numeric,
    lon numeric,
    parked boolean
);

ALTER TABLE ONLY cars
    ADD CONSTRAINT cars_pkey PRIMARY KEY (cid);
