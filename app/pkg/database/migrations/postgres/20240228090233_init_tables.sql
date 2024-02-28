-- +goose Up
-- +goose StatementBegin
CREATE TABLE resources (
  resource_id VARCHAR(255) NOT NULL PRIMARY KEY,
  resource_type VARCHAR(255) NOT NULL,
  resource_contact VARCHAR(255) NOT NULL UNIQUE,
  resource_slug VARCHAR(255) NOT NULL UNIQUE,
  resource_access_key VARCHAR(255) NOT NULL,
  access_granted BOOLEAN NOT NULL DEFAULT false,
  created_at DATETIME WITH TIME ZONE NOT NULL,
  updated_at DATETIME WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_resource_contact ON resources (resource_contact);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_resource_contact;
DROP TABLE resources;
-- +goose StatementEnd
