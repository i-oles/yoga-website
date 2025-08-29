ALTER DATABASE yoga SET timezone TO 'UTC';

CREATE TABLE classes
(
    id               uuid PRIMARY KEY,
    start_time       timestamptz NOT NULL,
    class_level      varchar(40) NOT NULL,
    class_name       varchar(40) NOT NULL,
    current_capacity integer     NOT NULL,
    max_capacity     integer     NOT NULL,
    location         varchar(40) NOT NULL
);

CREATE TYPE booking_operation AS ENUM ('create_booking', 'cancel_booking');

CREATE TABLE pending_operations
(
    id                 uuid PRIMARY KEY,
    class_id           uuid               NOT NULL,
    operation          booking_operation  NOT NULL,
    email              varchar(50)        NOT NULL,
    first_name         varchar(30)        NOT NULL,
    last_name          varchar(30),
    confirmation_token varchar(64) UNIQUE NOT NULL,
    created_at         timestamptz DEFAULT NOW()
);

ALTER TABLE pending_operations
    ADD FOREIGN KEY (class_id) REFERENCES classes (id);

CREATE TABLE confirmed_bookings
(
    id         uuid PRIMARY KEY,
    class_id   uuid        NOT NULL,
    first_name varchar(30) NOT NULL,
    last_name  varchar(30) NOT NULL,
    email      varchar(50) NOT NULL,
    created_at timestamptz DEFAULT (now())
);

ALTER TABLE confirmed_bookings
    ADD FOREIGN KEY (class_id) REFERENCES classes (id);

