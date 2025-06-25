INSERT INTO classes (day, datetime, level, place)
VALUES ('Wednesday', '2025-06-28 18:00', 'intermediate','Ogród Saski');

INSERT INTO classes (day, datetime, level, place)
VALUES ('Friday', '2025-06-30 18:00', 'beginner','Ogród Saski');

INSERT INTO classes (day, datetime, level, place)
VALUES ('Monday', '2025-05-15 18:00', 'beginner','Ogród Saski');

INSERT INTO classes (day, datetime, level, place)
VALUES ('Monday', ('2025-06-15 18:00' AT TIME ZONE 'Europe/Warsaw')::timestamptz, 'beginner','Ogród Saski');
