CREATE TABLE foo (
	id BIGINT PRIMARY KEY
);

CREATE TABLE bar
(
    foo_id  BIGINT NOT NULL,
    info    TEXT NOT NULL
);