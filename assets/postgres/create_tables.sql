ALTER DATABASE root SET timezone TO 'Europe/Warsaw';

-- CREATE TYPE class_level AS ENUM ('beginner', 'intermediate', 'advanced');

CREATE TABLE classes
(
    id             uuid PRIMARY KEY,
    day_of_week    varchar(20) NOT NULL,
    start_time     timestamp   NOT NULL,
    class_level    class_level NOT NULL,
    class_category varchar(30) NOT NULL,
    max_capacity   integer     NOT NULL DEFAULT 6,
    location       varchar(40) NOT NULL
);

-- CREATE TYPE booking_operation AS ENUM ('create_booking', 'cancel_booking');

CREATE TABLE pending_operations
(
    id               uuid PRIMARY KEY,
    class_id         uuid               NOT NULL,
    operation        booking_operation  NOT NULL,
    email            varchar(50)        NOT NULL,
    first_name       varchar(30)        NOT NULL,
    last_name        varchar(30),
    auth_token       varchar(64) UNIQUE NOT NULL,
    token_expires_at timestamp          NOT NULL,
    created_at       timestamp DEFAULT NOW()
);

CREATE TABLE confirmed_bookings
(
    id         uuid PRIMARY KEY,
    class_id   uuid        NOT NULL,
    first_name varchar(30) NOT NULL,
    last_name  varchar(30) NOT NULL,
    email      varchar(50) NOT NULL,
    created_at timestamp DEFAULT (now())
);

COMMENT ON COLUMN classes.max_capacity IS 'must be positive';

ALTER TABLE confirmed_bookings
    ADD FOREIGN KEY (class_id) REFERENCES classes (id);

ALTER TABLE pending_operations
    ADD FOREIGN KEY (class_id) REFERENCES classes (id);

CREATE UNIQUE INDEX idx_confirmed_bookings_unique
    ON confirmed_bookings (class_id, email);


