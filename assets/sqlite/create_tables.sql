CREATE TABLE classes
(
    id               TEXT PRIMARY KEY,
    start_time       TEXT NOT NULL,
    class_level      TEXT NOT NULL,
    class_name       TEXT NOT NULL,
    current_capacity INTEGER NOT NULL,
    max_capacity     INTEGER NOT NULL,
    location         TEXT NOT NULL
);

CREATE TABLE pending_operations
(
    id                 TEXT PRIMARY KEY,
    class_id           TEXT NOT NULL,
    operation          TEXT NOT NULL CHECK (operation IN ('create_booking', 'cancel_booking')),
    email              TEXT NOT NULL,
    first_name         TEXT NOT NULL,
    last_name          TEXT,
    confirmation_token TEXT UNIQUE NOT NULL,
    created_at         TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (class_id) REFERENCES classes (id)
);

CREATE TABLE confirmed_bookings
(
    id         TEXT PRIMARY KEY,
    class_id   TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name  TEXT NOT NULL,
    email      TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (class_id) REFERENCES classes (id)
);