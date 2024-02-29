-- +goose Up
-- +goose StatementBegin

CREATE TABLE hooks (
  id VARCHAR(100) NOT NULL PRIMARY KEY,
  hook_slug VARCHAR(120) NOT NULL,
  hook_secret VARCHAR(255) NOT NULL,
  allowed_logins TEXT,
);

CREATE TABLE events (

)

CREATE INDEX
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ;
-- +goose StatementEnd
