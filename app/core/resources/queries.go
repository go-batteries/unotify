package resources

const (
	CreateUserQuery = `
	INSERT INTO resources (
		resource_id
		,resource_type
		,resource_contact
		,resource_slug
		,resource_acccess_key
		,created_at
		,updated_at
	) VALUES (
		:resouce_id
		,'user'
		,:resource_contact
		,:resource_slug
		,:resource_acccess_key
		,:created_at
		,:updated_at
	)`

	FindUserByContact = `
	SELECT resource_id
		,resource_type
		,resource_contact
		,resource_slug
		,resource_access_key
		,created_at
	FROM resources
	WHERE resource_contact = $1
	`

	FindUserByID = `
	SELECT resource_id
		,resource_type
		,resource_contact
		,resource_slug
		,resource_access_key
		,created_at
	FROM resources
	WHERE resource_id = $1
	`
)
