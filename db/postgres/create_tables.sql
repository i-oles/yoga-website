CREATE TABLE classes
(
    id         bigserial PRIMARY KEY,
    day        varchar     NOT NULL,
    datetime   timestamptz NOT NULL,
    level      varchar     NOT NULL,
    spots_left integer     NOT NULL DEFAULT 6,
    place      varchar     NOT NULL
);

CREATE TABLE practitioners
(
    id         bigserial PRIMARY KEY,
    class_id   bigint      NOT NULL,
    nick       varchar,
    name       varchar     NOT NULL,
    last_name  varchar     NOT NULL,
    email      varchar     NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now()),
    updated_at timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON classes (id);

CREATE INDEX ON classes (datetime);

CREATE INDEX ON practitioners (class_id);

CREATE INDEX ON practitioners (email);

COMMENT ON COLUMN classes.spots_left IS 'must be positive';

ALTER TABLE practitioners
    ADD FOREIGN KEY (class_id) REFERENCES classes (id);