
-- +migrate Up

CREATE TABLE tasks (
    worker VARCHAR(10) NOT NULL,
    finished_flag CHAR(1) DEFAULT '0',
    task_name VARCHAR(25) NOT NULL,
    until DATETIME NOT NULL
);

-- +migrate Down

DROP TABLE tasks;