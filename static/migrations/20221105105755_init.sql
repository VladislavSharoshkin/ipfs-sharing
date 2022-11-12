-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE "contents" (
    "id"	INTEGER NOT NULL,
    "name"	TEXT NOT NULL,
    "cid"	TEXT NOT NULL,
    "parent_id"	INTEGER,
    "from" TEXT,
    PRIMARY KEY("id" AUTOINCREMENT)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE contents;
-- +goose StatementEnd
