package resources

import "time"

type User struct {
	ID        string `db:"resource_id"`
	Type      string `db:"resource_type"`
	Contact   string `db:"resource_contact"`
	Slug      string `db:"resource_slug"`
	AccessKey string `db:"resource_access_key"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
