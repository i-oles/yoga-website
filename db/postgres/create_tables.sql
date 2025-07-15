ALTER DATABASE root SET timezone TO 'Europe/Warsaw';

CREATE TABLE classes
(
    id         bigserial PRIMARY KEY,
    day        varchar     NOT NULL,
    datetime   timestamptz NOT NULL,
    level      varchar     NOT NULL,
    type       varchar     NOT NULL,
    spots_left integer     NOT NULL DEFAULT 6,
    place      varchar     NOT NULL
);

CREATE TABLE pending_bookings
(
    id         bigserial PRIMARY KEY,
    class_id   bigint             NOT NULL,
    email      varchar(60)        NOT NULL,
    name       varchar(30)        NOT NULL,
    last_name  varchar(40)        NOT NULL,
    token      varchar(64) UNIQUE NOT NULL,
    expires_at timestamp          NOT NULL,
    created_at timestamp DEFAULT NOW()
);

CREATE TABLE confirmed_bookings
(
    id         bigserial PRIMARY KEY,
    class_id   bigint      NOT NULL,
    name       varchar(60)     NOT NULL,
    last_name  varchar(30)     NOT NULL,
    email      varchar(40)     NOT NULL,
    created_at timestamp DEFAULT (now())
);

COMMENT ON COLUMN classes.spots_left IS 'must be positive';

ALTER TABLE confirmed_bookings
    ADD FOREIGN KEY (class_id) REFERENCES classes (id);

ALTER TABLE pending_bookings
    ADD FOREIGN KEY (class_id) REFERENCES classes (id);

CREATE UNIQUE INDEX idx_confirmed_bookings_unique
    ON confirmed_bookings (class_id, email);