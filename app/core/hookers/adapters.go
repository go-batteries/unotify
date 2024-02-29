package hookers

type RegisterHookRequest struct {
	Provider    string `json:"provider" form:"provider" validate:"required"`
	RepoID      string `json:"repo_id" form:"repo_id" validate:"required"`
	RepoPath    string `json:"repo_path" form:"repo_path" validate:"required"`
	ForceUpdate bool   `json:"-" form:"-" query:"-" validate:"-"`
}

type RegisterHookerResponse struct {
	Secret string `json:"secret,omitempty"`
}

type SearchHookerResponse struct {
	Secrets  string `json:"-"`
	RepoID   string `json:"repo_id,omitempty"`
	Provider string `json:"provider"`
}

type FindHookByProvider struct {
	Provider string `json:"provider" query:"provider" db:"provider" validate:"required"`
	RepoID   string `json:"repo_id" query:"repo_id" db:"repo_id"`
	Dive     bool   `json:"-" query:"-" db:"-" form:"-"`
}
