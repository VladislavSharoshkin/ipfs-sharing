-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE "contents" (
    "id"	INTEGER NOT NULL,
    "name"	TEXT NOT NULL,
    "cid"	TEXT NOT NULL,
    "parent_id"	INTEGER,
    "status" TEXT NOT NULL,
    "from" TEXT NOT NULL,
    "dir" TEXT NOT NULL,
    "created_at" TEXT NOT NULL,
    PRIMARY KEY("id" AUTOINCREMENT),
    UNIQUE(name, parent_id)
);

CREATE TABLE "messages" (
    "id"	INTEGER NOT NULL,
    "text"	TEXT NOT NULL,
    "from"	TEXT NOT NULL,
    "to" TEXT NOT NULL,
    "status" TEXT NOT NULL,
    "created_at" TEXT NOT NULL,
    PRIMARY KEY("id" AUTOINCREMENT)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE contents;
DROP TABLE messages;
-- +goose StatementEnd
